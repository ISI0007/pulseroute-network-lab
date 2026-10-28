"""Collect notices from the pinned local Go module cache and frontend runtime packages."""
import argparse, hashlib, json, pathlib, re
parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--module-cache',type=pathlib.Path,required=True)
parser.add_argument('--go-root',type=pathlib.Path,required=True)
args=parser.parse_args()
root=pathlib.Path(__file__).resolve().parents[1]
records=[]
sections=['PulseRoute third-party notices','',
 'PulseRoute itself is MIT licensed (see LICENSE). Upstream components retain their own licenses.',
 'Scope: all modules pinned in go.mod (including platform-specific/indirect modules), the Go runtime, and the four frontend production packages. Development-only npm tooling and OS/WebView2 components are not bundled here. This inventory does not claim every listed Go module is linked into both Windows executables.','']
def collect(component,version,folder):
    files=sorted(p for p in folder.iterdir() if p.is_file() and p.name.lower().startswith(('license','copying','notice')))
    if not files: raise RuntimeError(f'Missing upstream license notice: {component} {version}')
    record={'component':component,'version':version,'notices':[]}
    for p in files:
        content=p.read_bytes()
        record['notices'].append({'file':p.name,'sha256':hashlib.sha256(content).hexdigest()})
        sections.extend(['='*72,f'{component} {version} / {p.name}','='*72,content.decode('utf-8-sig',errors='replace'),''])
    records.append(record)
for module,version in re.findall(r'^\s*([\w.!~+/-]+\.[\w.!~+/-]+)\s+(v[^\s]+)',(root/'go.mod').read_text(),re.M):
    escaped=''.join('!'+c.lower() if c.isupper() else c for c in module)
    collect(module,version,args.module_cache/(escaped+'@'+version))
for name in ('react','react-dom','scheduler','lucide-react'):
    folder=root/'frontend/node_modules'/name
    package=json.loads((folder/'package.json').read_text())
    collect('npm:'+name,package['version'],folder)
collect('Go runtime',(args.go_root/'VERSION').read_text().splitlines()[0],args.go_root)
(root/'THIRD_PARTY_NOTICES.txt').write_text('\n'.join(line.rstrip() for line in '\n'.join(sections).splitlines())+'\n',encoding='utf-8')
(root/'docs/dependency-notices.json').write_text(json.dumps({'scope':'Pinned Go module license files, Go runtime and frontend production package notices; generated from existing installed upstream sources','components':records},indent=2)+'\n',encoding='utf-8')
print(f'PASS: collected notices for {len(records)} components')
