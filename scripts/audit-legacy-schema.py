"""Read-only ZIP schema evidence extractor. Never exports rows, credentials or .env files.

Usage: python scripts/audit-legacy-schema.py path/to/hms.zip
Outputs deterministic structural evidence in docs/legacy-schema/. PHP is inspected,
not executed. Migration operations are evidence, not a simulated final schema.
"""
import argparse
import csv
import collections
import hashlib
import json
import re
import zipfile
from pathlib import Path


def line_at(text, pos):
    return text.count('\n', 0, pos) + 1


def balanced(text, start, opening='(', closing=')'):
    depth, quote, escaped = 0, None, False
    for i in range(start, len(text)):
        c = text[i]
        if quote:
            if escaped:
                escaped = False
            elif c == '\\':
                escaped = True
            elif c == quote:
                quote = None
            continue
        if c in "'\"`":
            quote = c
        elif c == opening:
            depth += 1
        elif c == closing:
            depth -= 1
            if depth == 0:
                return text[start + 1:i], i + 1
    raise ValueError('Unbalanced source expression')


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('zip')
    ap.add_argument('--output', default='docs/legacy-schema')
    args = ap.parse_args()
    out = Path(args.output)
    out.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(args.zip) as archive:
        names = sorted(n for n in archive.namelist() if not n.endswith('/'))
        def read(name):
            return archive.read(name).decode('utf-8-sig', errors='replace')
        dump_path = next(n for n in names if n.endswith('/database/hms.sql'))
        root = dump_path.removesuffix('database/hms.sql')
        dump = read(dump_path)
        tables = {}
        for m in re.finditer(r'CREATE TABLE\s+`([^`]+)`\s*\(', dump):
            body, end = balanced(dump, m.end() - 1)
            ddl = dump[m.start():dump.index(';', end) + 1]
            columns = []
            for c in re.finditer(r'^\s*`([^`]+)`\s+(.+?)(?:,)?$', body, re.M):
                definition = c[2].rstrip(',')
                columns.append({'name': c[1], 'definition': definition,
                                'nullable': 'NOT NULL' not in definition,
                                'line': line_at(dump, m.end() + c.start(1))})
            tables[m[1]] = {'source': dump_path, 'line': line_at(dump, m.start()),
                            'columns': columns, 'create': ddl, 'alters': [], 'foreign_keys': []}
        for m in re.finditer(r'ALTER TABLE\s+`([^`]+)`.*?;', dump, re.S):
            assert m[1] in tables, m[1]
            tables[m[1]]['alters'].append({'line': line_at(dump, m.start()), 'sql': m[0]})
            for fk in re.finditer(r'CONSTRAINT `([^`]+)` FOREIGN KEY \(([^)]+)\) REFERENCES `([^`]+)` \(([^)]+)\)([^,;]*)', m[0]):
                tables[m[1]]['foreign_keys'].append({'name': fk[1], 'columns': re.findall(r'`([^`]+)`', fk[2]),
                    'target': fk[3], 'target_columns': re.findall(r'`([^`]+)`', fk[4]), 'actions': fk[5].strip()})
        dump_lines = dump.splitlines()
        for table in tables.values():
            for column in table['columns']:
                assert '`' + column['name'] + '`' in dump_lines[column['line'] - 1], 'Incorrect source line'
        assert sum(len(t['foreign_keys']) for t in tables.values()) == len(re.findall(r'FOREIGN KEY\s*\(', dump))
        migrations = []
        permission_config = read(root + 'config/permission.php')
        permission_names = dict(re.findall(r"['\"](roles|permissions|model_has_permissions|model_has_roles|role_has_permissions)['\"]\s*=>\s*['\"]([^'\"]+)", permission_config))
        for n in names:
            if not (n.startswith(root + 'database/migrations/') and n.endswith('.php')):
                continue
            src = read(n)
            up = src.split('function down', 1)[0]
            # Resolve this archive's explicitly configured Spatie table names.
            resolved = re.sub(r"\$tableNames\[['\"]([^'\"]+)['\"]\]", lambda m: "'" + permission_names[m[1]] + "'" if m[1] in permission_names else m[0], up)
            resolved = re.sub(r"\$columnNames\[['\"]model_morph_key['\"]\]", "'model_id'", resolved)
            operations = []
            for m in re.finditer(r"Schema::(create|table|rename|dropIfExists|drop)\(\s*['\"]([^'\"]+)['\"]", resolved):
                # Preserve all Blueprint statements for this block, not data writes.
                tail = resolved[m.start():]
                stop = tail.find('});')
                block = tail[:stop + 3] if stop >= 0 else tail.split(';', 1)[0]
                statements = [' '.join(x.split()) for x in re.findall(r'\$table\s*->.*?;', block, re.S)]
                operation = {'operation': m[1], 'table': m[2], 'line': line_at(resolved, m.start()), 'blueprint': statements}
                if m[1] == 'rename':
                    operation['rename_to'] = re.search(r",\s*['\"]([^'\"]+)", tail)[1]
                operations.append(operation)
            sql_ddl = []
            for m in re.finditer(r'(?:DB::statement|DB::unprepared)\(\s*([\'\"])(.*?)\1\s*\)', up, re.S):
                if re.match(r'\s*(ALTER|CREATE|DROP|RENAME)\s', m[2], re.I):
                    sql_ddl.append({'line': line_at(src, m.start()), 'sql': m[2]})
            migrations.append({'source': n, 'sha256': hashlib.sha256(archive.read(n)).hexdigest(),
                'operations': operations, 'raw_ddl': sql_ddl,
                'seed_calls': re.findall(r'(\w+Seeder)::class', up),
                'needs_manual_review': bool(re.search(r'DB::|Schema::(?:hasColumn|hasTable)|config\(', up))})
        models = []
        for n in names:
            if not (n.startswith(root + 'app/Models/') and n.endswith('.php')):
                continue
            src = read(n)
            cls = re.search(r'\bclass\s+(\w+)', re.sub(r'/\*.*?\*/', '', src, flags=re.S))
            explicit = re.search(r"\$table\s*=\s*['\"]([^'\"]+)", src)
            props = {}
            for p in ('fillable', 'casts', 'rules', 'hidden'):
                m = re.search(r'\$' + p + r'\s*=\s*\[', src)
                if m:
                    props[p] = balanced(src, m.end() - 1, '[', ']')[0].strip()
            rels = []
            methods = list(re.finditer(r'(?:public|protected)\s+(?:static\s+)?function\s+(\w+)\s*\(', src))
            for i, method in enumerate(methods):
                body = src[method.end():methods[i + 1].start() if i + 1 < len(methods) else len(src)]
                for r in re.finditer(r'\$this->(belongsToMany|belongsTo|hasMany|hasOne|morphTo|morphMany|morphOne|morphToMany|hasManyThrough|hasOneThrough)\s*\(', body):
                    rels.append({'method': method[1], 'kind': r[1], 'arguments': balanced(body, r.end() - 1)[0].strip(),
                                 'line': line_at(src, method.end() + r.start())})
            constants = []
            for m in re.finditer(r'\bconst\s+([A-Z_][A-Z_0-9]*)\s*=\s*(.*?);', src, re.S):
                constants.append({'name': m[1], 'value': m[2].strip(), 'line': line_at(src, m.start())})
            inferred = {'User': 'users', 'Permission': 'permissions', 'EmailTemplate': 'email_templates',
                        'PurchaseMedicine': 'purchase_medicines', 'PurchasedMedicine': 'purchased_medicines'}
            models.append({'source': n, 'class': cls[1] if cls else None, 'explicit_table': explicit[1] if explicit else None,
                           'resolved_table': explicit[1] if explicit else inferred.get(cls[1] if cls else ''),
                           'properties': props, 'relationships': rels, 'constants': constants})
        evidence_files = []
        for n in names:
            if n.endswith('.php') and any(n.startswith(root + p) for p in ['app/Repositories/', 'app/Http/Controllers/', 'app/Http/Requests/', 'routes/']):
                src = read(n)
                evidence_files.append({'source': n, 'sha256': hashlib.sha256(archive.read(n)).hexdigest(),
                    'methods': [{'name': m[1], 'line': line_at(src, m.start())} for m in re.finditer(r'public\s+(?:static\s+)?function\s+(\w+)\s*\(', src)],
                    'referenced_models': sorted(set(re.findall(r'use App\\Models\\(\w+)', src))),
                    'transaction_lines': [line_at(src, m.start()) for m in re.finditer(r'DB::(?:transaction|beginTransaction|commit|rollBack)', src)]})
        pg = []
        for p in sorted(Path('db/migrations').glob('*.sql')):
            src = p.read_text(encoding='utf-8-sig')
            # Exact source DDL; these operations are NOT a queried deployed schema.
            creates = []
            for m in re.finditer(r'CREATE TABLE\s+(?:IF NOT EXISTS\s+)?("[^"]+"|\w+)\s*\(', src, re.I):
                body, end = balanced(src, m.end() - 1)
                creates.append({'table': m[1].strip('"'), 'line': line_at(src, m.start()), 'definition': body})
            pg.append({'source': p.as_posix(), 'creates': creates,
                       'alters': [m[0] for m in re.finditer(r'ALTER TABLE\b[^;]+;', src, re.I)],
                       'indexes': [m[0] for m in re.finditer(r'CREATE\s+(?:UNIQUE\s+)?INDEX\b[^;]+;', src, re.I)]})
        release_ddl = []
        for n in names:
            if n.startswith(root + 'database/releases/') and n.endswith('.sql'):
                src = read(n)
                release_ddl.append({'source': n, 'sha256': hashlib.sha256(archive.read(n)).hexdigest(),
                    'ddl': [m[0] for m in re.finditer(r'^(?:CREATE TABLE|ALTER TABLE|DROP TABLE|RENAME TABLE)\b.*?;', src, re.M | re.S)]})
        result = {'source_dump': dump_path, 'dump_sha256': hashlib.sha256(archive.read(dump_path)).hexdigest(),
            'dump_header': re.findall(r'^-- (?:Generation Time|Server version|Database).*', dump, re.M),
            'scope': 'DDL and static source metadata only; no data rows or environment secrets; no PHP execution.',
            'tables': tables, 'migrations': migrations, 'models': models, 'workflow_sources': evidence_files,
            'permission_table_config': permission_names, 'release_ddl': release_ddl,
            'postgres_migrations': pg}
        (out / 'inventory.json').write_text(json.dumps(result, indent=2, ensure_ascii=False) + '\n', encoding='utf-8')
        schema = ['-- Evidence only. Original MySQL DDL, not a PostgreSQL migration. No INSERT statements.']
        schema.extend(t['create'] for t in tables.values())
        schema.extend(a['sql'] for t in tables.values() for a in t['alters'])
        (out / 'original-mysql-schema.sql').write_text('\n\n'.join(schema) + '\n', encoding='utf-8')
        md = ['# Original ZIP database dictionary', '',
              'Generated by `scripts/audit-legacy-schema.py`. Snapshot from `hms/database/hms.sql`; migration differences are tracked separately. Not the live demo database. No source records are exported.', '',
              'Column definitions preserve MySQL types, nullability and defaults. Keys, indexes, auto-increment and referential actions appear after each table. Model associations are **application evidence**, not proof of a database foreign key.', '']
        for name, table in tables.items():
            md += [f'## {name}', '', f'Source: `{dump_path}:{table["line"]}`', '', '| Column | MySQL definition |', '| --- | --- |']
            md += [f'| `{c["name"]}` | `{c["definition"].replace("|", "&#124;")}` |' for c in table['columns']]
            md += ['', '```sql', '\n'.join(a['sql'] for a in table['alters']), '```', '']
            matched = [x for x in models if x['resolved_table'] == name]
            for model in matched:
                md += [f'Model: `{model["source"]}`', '']
                md += [f'- `{r["method"]}` → `{r["kind"]}({r["arguments"]})` (line {r["line"]})' for r in model['relationships']]
                md += ['']
        (out / 'TABLES.md').write_text('\n'.join(md).rstrip() + '\n', encoding='utf-8')
        chronology = ['# Migration evidence', '',
            'All application migration up-directions, in filename order. Blueprint excerpts are structural evidence, not an executed final schema. Seeder call names are listed without seed contents. Configured permission table names resolve through `hms/config/permission.php`.', '']
        state, unhandled = {}, []
        column_methods = set('bigincrements increments string text longtext mediumtext integer biginteger unsignedbiginteger unsignedinteger smallinteger tinyinteger unsignedtinyinteger boolean date datetime timestamp time decimal double float json uuid char binary enum unsignedsmallinteger foreignid'.split())
        for migration in migrations:
            chronology += [f'## {Path(migration["source"]).name}', '', f'Source: `{migration["source"]}`', '']
            for op in migration['operations']:
                chronology += [f'- `{op["operation"]}` `{op["table"]}` at line {op["line"]}' + (f' → `{op["rename_to"]}`' if 'rename_to' in op else '')]
                if op['blueprint']:
                    chronology += ['', '```php', '\n'.join(op['blueprint']), '```', '']
                t = op['table']
                if op['operation'] == 'rename':
                    state[op['rename_to']] = state.pop(t, {})
                    continue
                if op['operation'] in ('drop', 'dropIfExists'):
                    state.pop(t, None)
                    continue
                if op['operation'] == 'create':
                    state[t] = {}
                for stmt in op['blueprint']:
                    m = re.match(r'\$table\s*->(\w+)\((.*)', stmt)
                    if not m:
                        unhandled.append({'source': migration['source'], 'statement': stmt})
                        continue
                    method, arg = m[1].lower(), m[2]
                    col = re.match(r"['\"]([^'\"]+)['\"]", arg)
                    cols = state.setdefault(t, {})
                    if method in column_methods and col:
                        cols[col[1]] = migration['source']
                    elif method == 'id':
                        cols[col[1] if col else 'id'] = migration['source']
                    elif method in ('timestamps', 'nullabletimestamps'):
                        cols.update(created_at=migration['source'], updated_at=migration['source'])
                    elif method in ('remembertoken', 'softdeletes'):
                        cols['remember_token' if method == 'remembertoken' else 'deleted_at'] = migration['source']
                    elif method in ('morphs', 'nullablemorphs') and col:
                        cols.update({col[1] + '_id': migration['source'], col[1] + '_type': migration['source']})
                    elif method == 'dropcolumn':
                        for c in re.findall(r"['\"]([^'\"]+)['\"]", arg):
                            cols.pop(c, None)
                    elif method == 'renamecolumn':
                        cs = re.findall(r"['\"]([^'\"]+)['\"]", arg)
                        cols.pop(cs[0], None)
                        cols[cs[1]] = migration['source']
                    elif method not in ('primary', 'index', 'unique', 'foreign', 'dropforeign', 'dropindex', 'dropunique'):
                        unhandled.append({'source': migration['source'], 'statement': stmt})
            if migration['seed_calls']:
                chronology += ['Seeder references: ' + ', '.join(f'`{s}`' for s in migration['seed_calls']), '']
            if not migration['operations']:
                chronology += ['No literal/resolved Schema operation extracted; see source for data/seeder-only work.', '']
        (out / 'MIGRATIONS.md').write_text('\n'.join(chronology), encoding='utf-8')
        differences = {'limitation': 'Column-name reconciliation only; no execution or type/default/conditional-path equivalence claim.',
            'dump_only_tables': sorted(set(tables) - set(state)), 'migration_only_tables': sorted(set(state) - set(tables)),
            'unhandled_blueprint_statements': unhandled, 'column_differences': []}
        for name, cols in state.items():
            if name in tables:
                dump_cols = {c['name'] for c in tables[name]['columns']}
                if set(cols) != dump_cols:
                    differences['column_differences'].append({'table': name, 'dump_only': sorted(dump_cols - set(cols)),
                                                             'migration_only': sorted(set(cols) - dump_cols)})
        (out / 'migration-reconciliation.json').write_text(json.dumps(differences, indent=2) + '\n', encoding='utf-8')
        # All FK edges, including self-links, retain exact direction/action evidence.
        with (out / 'foreign-keys.csv').open('w', encoding='utf-8', newline='') as f:
            writer = csv.writer(f)
            writer.writerow(['table', 'columns', 'referenced_table', 'referenced_columns', 'constraint', 'actions'])
            for name, table in tables.items():
                for fk in table['foreign_keys']:
                    assert fk['target'] in tables
                    assert set(fk['columns']) <= {c['name'] for c in table['columns']}
                    assert set(fk['target_columns']) <= {c['name'] for c in tables[fk['target']]['columns']}
                    writer.writerow([name, ','.join(fk['columns']), fk['target'], ','.join(fk['target_columns']), fk['name'], fk['actions']])
        spec_path = Path('docs/legacy-schema/mapping-spec.tsv')
        if spec_path.exists():
            specs = list(csv.DictReader(spec_path.open(encoding='utf-8'), delimiter='\t'))
            mapped = [t for row in specs for t in row['legacy_tables'].split(',')]
            assert collections.Counter(mapped) == collections.Counter(tables.keys()), 'Every legacy table must be mapped exactly once'
            pg_tables = {c['table'] for p in pg for c in p['creates']}
            assert {t for row in specs for t in row['postgres_candidates'].split(',') if t != '-'} <= pg_tables
            mapping = ['# Original → PostgreSQL table coverage', '',
                'Every original dump table appears exactly once. Targets are **candidate storage counterparts**, not verified field or API parity. A dash means no dedicated counterpart was identified in the 48 checked-in migrations. Grouping is architectural interpretation; original facts are in TABLES.md and inventory.json.', '',
                '| Domain | Original table | PostgreSQL candidates | Required reconciliation |', '| --- | --- | --- | --- |']
            by_table = {}
            for row in specs:
                for name in row['legacy_tables'].split(','):
                    by_table[name] = row
                    mapping.append(f'| {row["domain"]} | `{name}` | {row["postgres_candidates"]} | {row["assessment"]} |')
            (out / 'POSTGRES-MAPPING.md').write_text('\n'.join(mapping) + '\n', encoding='utf-8')
            with (out / 'field-parity-register.csv').open('w', encoding='utf-8', newline='') as f:
                writer = csv.writer(f)
                writer.writerow(['domain', 'legacy_table', 'legacy_column', 'legacy_definition', 'candidate_target_tables',
                                 'target_column_or_transform', 'api_field', 'ui_workflow', 'verification_status', 'evidence'])
                for name, table in tables.items():
                    for col in table['columns']:
                        writer.writerow([by_table[name]['domain'], name, col['name'], col['definition'], by_table[name]['postgres_candidates'],
                            '', '', '', 'pending_field_contract_review', f'{dump_path}:{col["line"]}'])
        summary = {'tables': len(tables), 'columns': sum(len(t['columns']) for t in tables.values()),
            'foreign_keys': sum(len(t['foreign_keys']) for t in tables.values()), 'migrations': len(migrations),
            'models': len(models), 'model_relationships': sum(len(m['relationships']) for m in models),
            'workflow_files_indexed': len(evidence_files), 'postgres_migrations': len(pg),
            'postgres_create_tables': sum(len(p['creates']) for p in pg)}
        print(json.dumps(summary, indent=2))


if __name__ == '__main__':
    main()
