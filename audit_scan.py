import os, re
root='.'
include_dirs=['daemon','src','vscode-extension','docs']
pats=[r'C:\\', r'/home/', r'/Users/', r'fmt\.Println\(', r'log\.Println\(', r'console\.log\(', r'package\.json', r'npm', r'node_modules', r'go\.mod', r'poetry', r'pyproject', r'\.bashrc', r'24', r'500', r'5000']
for base,_,fs in os.walk(root):
    rel=os.path.relpath(base,root)
    if rel=='.':
        pass
    elif rel.split(os.sep)[0] not in include_dirs:
        continue
    for f in fs:
        if not f.endswith(('.go','.ts','.js','.md','.json','.yaml','.yml','.toml','.ps1','.sh')):
            continue
        path=os.path.join(base,f)
        try:
            text=open(path,encoding='utf-8').read()
        except Exception:
            continue
        hits=[pat for pat in pats if re.search(pat,text)]
        if hits:
            print(path)
            for pat in hits:
                print('  ',pat)
