"""Build source-reference appendices and a documentation-only portable handoff."""
from pathlib import Path
import hashlib
import json
import re
import zipfile

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / 'docs/builder-handoff'
OUT.mkdir(parents=True, exist_ok=True)

def read(path):
    return (ROOT / path).read_text(encoding='utf-8')

def write(name, content):
    (OUT / name).write_text(content, encoding='utf-8')

catalog = json.loads(read('apps/web/src/lib/legacy-catalog.json'))
schema = json.loads(read('docs/legacy-schema/inventory.json'))
matrix = read('docs/module-audit/ACCEPTANCE-MATRIX.md')
rows = []
for line in matrix.splitlines():
    if line.startswith('| ') and not line.startswith('| Module') and not line.startswith('| ---'):
        cells = [v.strip() for v in line.strip('|').split('|')]
        if len(cells) == 4:
            rows.append(cells)
parts = ['# Workspace-by-workspace implementation reference\n',
         'This expands the source-audit acceptance matrix into individual work packets. The audit observations are historical; the repair notes below supersede only their named findings. Every workspace also requires the shared action/role/ownership/failure tests in SECURITY-AND-ACCEPTANCE.md. A generic list is not completion.\n']
for name, source, checkpoint, missing in rows:
    parts += [f'## {name}\n', f'**Original obligation / workflow:** {source}.\n',
              f'**Audit checkpoint:** {checkpoint}.\n', f'**Required implementation and acceptance:** {missing}.\n',
              '**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.\n']
parts += ['## Corrections after the historical matrix\n',
          'General-settings proxy and transactional bulk updates, CMS permission advertising, admin aggregate doctor absences and inventory movements are repaired. Inventory category/item create and receive/issue persistence are repaired. See ../module-audit/REPAIRS.md and README.md. Full forms, remaining inventory actions and all-module parity remain open.\n',
          'Attendance is demo-derived, not ZIP-derived. Add-on behavior and unsafe historical role/delete behavior require explicit design decisions. Public CMS, administration, personal staff and patient views must remain distinct even when they use the same underlying entity.\n']
write('WORKSPACES.md', '\n'.join(parts))

parts = ['# Original screen field inventory\n',
         f'This appendix contains all {len(catalog)} records in the committed source-derived catalog. It is not a claim that every dynamic/conditional field was extracted. Read MODEL-RULES.md and table definitions for server requirements. `extracted: false` records are not source-confirmed forms. Current target overrides and newer attendance screens are supplied separately in reference/legacy.ts; those overrides are implementation evidence, not proof of original rules.\n']
for screen in catalog:
    parts += [f"## {screen['title']} (`{screen['id']}`)\n",
              f"Group: {screen.get('group', '')}. Extraction flag: `{screen.get('extracted', 'unspecified')}`.\n",
              f"Original source pointer: `{screen.get('source', '')}`. Table pointer: `{screen.get('tableSource', '')}`. These provenance paths need not exist in the new repository.\n",
              '| Field key | Label | UI type | Required flag |\n| --- | --- | --- | --- |']
    for field in screen.get('fields', []):
        vals = [str(field.get(k, '')).replace('|', '\\|').replace('\n', ' ') for k in ('key', 'label', 'type', 'required')]
        parts.append('| ' + ' | '.join(vals) + ' |')
    parts.append('\nTable columns: ' + json.dumps(screen.get('columns', []), ensure_ascii=False) + '.\n')
write('WORKSPACE-FIELDS.md', '\n'.join(parts))

manifest = {
    'prepared': '2026-10-08', 'implementation_baseline': '72c00a7',
    'original_tables': len(schema['tables']),
    'original_columns': sum(len(t['columns']) for t in schema['tables'].values()),
    'original_foreign_keys': sum(len(t['foreign_keys']) for t in schema['tables'].values()),
    'catalog_screens': len(catalog), 'module_families': len(rows),
    'source_dump_sha256': schema['dump_sha256'],
    'scope': 'Documentation and source-derived metadata only; no Laravel runtime, database rows or environment credentials.'
}
write('handoff-manifest.json', json.dumps(manifest, indent=2) + '\n')

# Check local Markdown file links in the new handoff. Original source pointers are code text.
for doc in OUT.glob('*.md'):
    for target in re.findall(r'\]\(([^)]+)\)', doc.read_text(encoding='utf-8')):
        if '://' in target or target.startswith('#'):
            continue
        assert (doc.parent / target.split('#')[0]).exists(), (doc.name, target)
assert manifest['original_tables'] == 141 and manifest['original_columns'] == 1202
assert manifest['original_foreign_keys'] == 144 and manifest['catalog_screens'] == 111
assert not re.search(r'^\s*INSERT\s+INTO\b', read('docs/legacy-schema/original-mysql-schema.sql'), re.M | re.I)

bundle = ROOT / '.local/handoff/HMS-AI-BUILDER-HANDOFF.zip'
bundle.parent.mkdir(parents=True, exist_ok=True)
files = sorted(p for p in (ROOT / 'docs').rglob('*') if p.is_file() and p.suffix.lower() in {'.md', '.json', '.sql', '.csv', '.tsv', '.yaml', '.yml'})
checksums = {}
with zipfile.ZipFile(bundle, 'w', zipfile.ZIP_DEFLATED) as archive:
    archive.writestr('START-HERE.md', '# HMS AI builder handoff\n\nRead [the current handoff](docs/builder-handoff/README.md) first. Older reports are historical; the current handoff explains subsequent repairs and remaining gaps.\n')
    for path in files:
        name = path.relative_to(ROOT).as_posix()
        content = path.read_bytes()
        archive.writestr(name, content)
        checksums[name] = hashlib.sha256(content).hexdigest()
    for name in ('legacy-catalog.json', 'prescription-fields.json', 'legacy.ts'):
        path = ROOT / 'apps/web/src/lib' / name
        content = path.read_bytes()
        archive.writestr('reference/' + name, content)
        checksums['reference/' + name] = hashlib.sha256(content).hexdigest()
    archive.writestr('SHA256-MANIFEST.json', json.dumps(checksums, indent=2))
with zipfile.ZipFile(bundle) as archive:
    assert archive.testzip() is None
    for name, digest in checksums.items():
        assert hashlib.sha256(archive.read(name)).hexdigest() == digest
print(json.dumps({'manifest': manifest, 'archive': str(bundle), 'files': len(checksums), 'bytes': bundle.stat().st_size}, indent=2))
