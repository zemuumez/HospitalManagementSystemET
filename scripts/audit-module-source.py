"""Static navigation/guard/validation evidence, not an executable PHP audit.
Run after audit-legacy-schema.py. No source data, environment or credentials exported.
"""
import argparse
import collections
import importlib.util
import json
import re
import sys
import zipfile
from pathlib import Path

sys.dont_write_bytecode = True
spec = importlib.util.spec_from_file_location('schema_audit', Path(__file__).with_name('audit-legacy-schema.py'))
helper = importlib.util.module_from_spec(spec)
spec.loader.exec_module(helper)


def uncomment(s):
    # Preserve character offsets and quoted strings when removing comments.
    return re.sub(r'''('(?:\\.|[^'\\])*'|"(?:\\.|[^"\\])*")|(/\*.*?\*/|//[^\n]*|\#[^\n]*)''',
                  lambda m: m[1] if m[1] is not None else re.sub(r'[^\n]', ' ', m[0]), s, flags=re.S)


def evidence_lines(s, pattern, offset=0):
    return [{'line': offset + i, 'text': line.strip()} for i, line in enumerate(s.splitlines(), 1)
            if re.search(pattern, line)]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('zip')
    args = ap.parse_args()
    out = Path('docs/module-audit')
    out.mkdir(exist_ok=True)
    inventory = json.loads(Path('docs/legacy-schema/inventory.json').read_text(encoding='utf-8'))
    routes, classes, permission_seeds = [], [], []
    with zipfile.ZipFile(args.zip) as z:
        for name in sorted(z.namelist()):
            if not name.endswith('.php'):
                continue
            if name in ('hms/database/seeders/AssignDefaultRoleToUserSeeder.php', 'hms/database/seeders/AddLabTechnicianPermissionSeeder.php', 'hms/database/seeders/PermissionTableSeeder.php'):
                seed = uncomment(z.read(name).decode('utf-8-sig', errors='replace'))
                permission_seeds.append({'source': name,
                    'permission_names': sorted(set(re.findall(r"['\"](manage_[a-z_]+)['\"]", seed))),
                    'role_lookups': re.findall(r"\$(\w+)\s*=\s*Department::whereName\(['\"]([^'\"]+)['\"]\)", seed),
                    'named_permission_sets': [{'variable': m[1], 'permissions': re.findall(r"['\"](manage_[a-z_]+)['\"]", m[2])} for m in re.finditer(r'\$(\w+)\s*=\s*Permission::whereIn\((.*?)\)->get\(\)',seed,re.S)],
                    'grant_calls': evidence_lines(seed,r'givePermissionTo\(|Permission::all\(')})
            if not name.startswith(('hms/routes/', 'hms/app/Http/Controllers/', 'hms/app/Http/Requests/', 'hms/app/Repositories/')):
                continue
            source = z.read(name).decode('utf-8-sig', errors='replace')
            clean = uncomment(source)
            if name.startswith('hms/routes/'):
                groups = []
                for g in re.finditer(r'Route::(?:middleware|group|prefix|name)\b[^;{}]*function\s*\([^)]*\)\s*\{', clean):
                    _, end = helper.balanced(clean, g.end()-1, '{', '}')
                    groups.append((g.start(), end, ' '.join(clean[g.start():g.end()-1].split())))
                for r in re.finditer(r'Route::(get|post|put|patch|delete|resource|apiResource|any|match|options)\s*\(', clean):
                    _, end = helper.balanced(clean, r.end()-1)
                    stop = clean.find(';', end)
                    statement = ' '.join(clean[r.start():stop+1].split()) if stop >= 0 else clean[r.start():end]
                    controller = re.search(r'\[\s*([\w\\]+)::class\s*,\s*[\'"]([^\'"]+)', statement)
                    resource = re.search(r",\s*([\w\\]+)::class", statement)
                    path = re.match(r"Route::\w+\(\s*['\"]([^'\"]+)", statement)
                    routes.append({'source': name, 'line': helper.line_at(source, r.start()), 'kind': r[1],
                        'literal_path': path[1] if path else None, 'declaration': statement,
                        'controller': controller[1] if controller else resource[1] if resource else None,
                        'action': controller[2] if controller else 'resource expansion (check only/except)' if resource else 'closure/dynamic',
                        'enclosing_route_groups': [text for start, finish, text in groups if start < r.start() < finish]})
                continue
            methods = []
            for m in re.finditer(r'(?:public|protected|private)\s+(?:static\s+)?function\s+(\w+)\s*\([^)]*\)[^{;]*\{', clean):
                body, end = helper.balanced(clean, m.end()-1, '{', '}')
                start_line = helper.line_at(source, m.end())-1
                methods.append({'name': m[1], 'line': helper.line_at(source, m.start()),
                    'guards_validation': evidence_lines(body, r'\bif\s*\(|authorize\(|hasRole\(|hasPermission|can\(|validate\(|::\$rules|return\s+(true|false)', start_line),
                    'mutations_transactions': evidence_lines(body, r'->(?:create|update|delete|forceDelete|save|store|sync|attach|detach)\(|::(?:create|destroy|transaction|beginTransaction|commit|rollBack)\(', start_line),
                    'view_names': re.findall(r"view\(\s*['\"]([^'\"]+)", body),
                    'model_calls': sorted(set(re.findall(r'\b([A-Z]\w*)::(\w+)\(', body))),
                    'request_type': ' '.join(clean[m.start():m.end()-1].split())})
            classes.append({'source': name, 'methods': methods,
                'imports': re.findall(r'^use\s+(App\\[^;]+);', clean, re.M),
                'middleware': evidence_lines(clean, r'->middleware\(')})
    current = []
    for path in sorted(Path('services/api/internal').rglob('*.go')):
        src = path.read_text(encoding='utf-8')
        current.append({'source': path.as_posix(),
            'tests': [{'name': m[1], 'line': helper.line_at(src, m.start())} for m in re.finditer(r'func (Test\w+)\(', src)],
            'auth_guards': evidence_lines(src, r'\.Can\(|\.Role\s*[!=]=|AuthorizeClinicalRecord\('),
            'route_conditions': evidence_lines(src, r'r\.URL\.Path ==|strings.HasPrefix\(r.URL.Path|case "[a-z][\w-]*":'),
            'test_skips': evidence_lines(src, r'\bt\.Skip')})
    frontend = []
    for path in sorted(Path('apps/web/src').rglob('*.tsx')):
        src = path.read_text(encoding='utf-8')
        if '/components/' not in path.as_posix() and '/app/(hospital)/' not in path.as_posix():
            continue
        frontend.append({'source': path.as_posix(),
            'api_calls': evidence_lines(src, r'\bfetch\(|\bapi(?:<.*>)?\('),
            'persistence_or_fixture_signals': evidence_lines(src, r'localStorage|sessionStorage|initial[A-Z]|demo[A-Z]|fallback|preview',),
            'mutation_methods': evidence_lines(src, r'method:\s*["\'](?:POST|PUT|PATCH|DELETE)'),
            'failure_handlers': evidence_lines(src, r'catch\s*[({]'),
            'component_imports': re.findall(r'from ["\']([^"\']+)["\']', src)})
    result = {'scope': 'Static source indexes. Guard expressions are candidates, not evaluated authorization. Resource routes are not expanded. Transitive imports and dynamic calls require review.',
              'legacy_routes': routes, 'legacy_workflow_classes': classes, 'legacy_permission_seed_evidence': permission_seeds, 'current_go': current, 'current_frontend': frontend}
    (out/'source-evidence.json').write_text(json.dumps(result, indent=2, ensure_ascii=False)+'\n', encoding='utf-8')
    md = ['# Original action and route-guard register', '', 'Static ZIP evidence. Resource declarations retain only/except modifiers; enclosing groups show role/permission middleware. Controller/request/row ownership still needs evaluation. Route order is preserved within files.', '', '| Source | Declaration | Enclosing groups |', '| --- | --- | --- |']
    for r in routes:
        escape = lambda s: s.replace('|', '&#124;').replace('`', '')
        md.append(f'| `{r["source"]}:{r["line"]}` | `{escape(r["declaration"])}` | {escape(" / ".join(r["enclosing_route_groups"]))} |')
    (out/'LEGACY-ACTIONS.md').write_text('\n'.join(md)+'\n', encoding='utf-8')
    model_by_name = {m['class']: m for m in inventory['models']}
    md = ['# Original controller workflow register', '', 'Rules/guards below are static source excerpts. They are not proof that all runtime paths were executed or that legacy behavior is safe to copy. Read the complete method for context. Models and full validation arrays are in ../legacy-schema/inventory.json.', '']
    for cls in classes:
        md += [f'## {cls["source"]}', '']
        referenced = [model_by_name.get(i.split('\\')[-1]) for i in cls['imports'] if i.startswith('App\\Models\\')]
        md += [('Models: '+', '.join(f'`{m["class"]}` → `{m["resolved_table"]}`' for m in referenced if m)).rstrip(), '']
        for method in cls['methods']:
            md += [f'### {method["name"]} (line {method["line"]})', '', f'`{method["request_type"]}`', '']
            for label, key in [('Guard/validation candidates', 'guards_validation'), ('Write/transaction candidates', 'mutations_transactions')]:
                if method[key]:
                    md += [label+':', '', '```php', '\n'.join(f'{v["line"]}: {v["text"]}' for v in method[key]), '```', '']
            if method['view_names']:
                md += ['Views: '+', '.join(f'`{v}`' for v in method['view_names']), '']
    (out/'LEGACY-WORKFLOWS.md').write_text('\n'.join(md).rstrip()+'\n', encoding='utf-8')
    md = ['# Original model validation and state constants', '',
          'Exact static model declarations. FormRequest overrides, repository checks and database constraints can add or change these rules. These expressions were not executed.', '']
    for model in inventory['models']:
        md += [f'## {model["class"]}', '', f'Source: `{model["source"]}`; table: `{model["resolved_table"]}`', '']
        rules = model['properties'].get('rules')
        md += ['```php', rules or '// No static rules array extracted from this model.', '```', '']
        for const in model['constants']:
            md += ['```php', f'const {const["name"]} = {const["value"]};', '```', '']
    (out/'MODEL-RULES.md').write_text('\n'.join(md).rstrip()+'\n', encoding='utf-8')
    md = ['# Frontend integration source register', '',
          'All inspected component/hospital-page files. A fetch count does not establish working integration; local storage for theme/language is harmless. Signals require action-level review, and helper/transitive API calls may live in imported .ts files.', '',
          '| File | Direct API call sites | Explicit mutation method sites | Storage/fixture/preview signals |', '| --- | --- | --- | --- |']
    for entry in frontend:
        md.append(f'| `{entry["source"]}` | {len(entry["api_calls"])} | {len(entry["mutation_methods"])} | {len(entry["persistence_or_fixture_signals"])} |')
    for entry in frontend:
        if entry['api_calls'] or entry['persistence_or_fixture_signals']:
            md += ['', f'## {entry["source"]}', '']
            for label,key in [('Direct calls','api_calls'),('Persistence/fixture signals','persistence_or_fixture_signals')]:
                md += [label, '', '```tsx', '\n'.join(f'{r["line"]}: {r["text"]}' for r in entry[key]), '```', '']
    (out/'FRONTEND-INTEGRATION.md').write_text('\n'.join(md).rstrip()+'\n', encoding='utf-8')
    src = Path('services/api/internal/domain/hospital.go').read_text()
    can, advertised = src.split('func (a Actor) Permissions()',1)
    policy = {}
    for m in re.finditer(r'case ([^:]+):\s*return ([^\n]+)',can):
        for permission in re.findall(r'"([^"]+)"',m[1]):
            policy[permission] = re.findall(r'a.Role == "([^"]+)"',m[2])
    published = re.findall(r'"([a-z_]+\.[a-z_]+)"',advertised.split('type PatientInput',1)[0])
    summary = {'limitation':'Static Can switch extraction; row ownership, direct Role checks and effective HTTP authorization remain separate.',
               'permission_roles':policy,'not_advertised':sorted(set(policy)-set(published)),
               'duplicate_advertisements':sorted(k for k,v in collections.Counter(published).items() if v>1)}
    (out/'CURRENT-ROLE-POLICY.json').write_text(json.dumps(summary,indent=2)+'\n',encoding='utf-8')
    print(json.dumps({'legacy_route_declarations':len(routes), 'legacy_workflow_classes':len(classes),
        'legacy_methods':sum(len(c['methods']) for c in classes),'current_go_files':len(current),'frontend_files':len(frontend)}))


if __name__ == '__main__':
    main()
