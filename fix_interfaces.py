import os
import re

replacements = {
    r'\.User\.Client\.': '.User.Client().',
    r'\.User\.Session\.': '.User.Session().',
    r'\.Audit\.Log\.': '.Audit.Log().',
    r'\.Audit\.History\.': '.Audit.History().',
    r'\.Audit\.Metric\.': '.Audit.Metric().',
    r'\.Audit\.SystemError\.': '.Audit.SystemError().',
    r'\.Authz\.Role\.': '.Authz.Role().',
    r'\.Authz\.Permission\.': '.Authz.Permission().',
    r'\.Authz\.Policy\.': '.Authz.Policy().',
    r'\.Authz\.Relation\.': '.Authz.Relation().',
    r'\.Authz\.Scope\.': '.Authz.Scope().',
}

def fix_files(directory):
    for root, dirs, files in os.walk(directory):
        for file in files:
            if file.endswith('.go'):
                filepath = os.path.join(root, file)
                with open(filepath, 'r') as f:
                    content = f.read()
                
                new_content = content
                for pattern, repl in replacements.items():
                    new_content = re.sub(pattern, repl, new_content)
                
                if new_content != content:
                    with open(filepath, 'w') as f:
                        f.write(new_content)
                    print(f"Fixed: {filepath}")

fix_files('internal/controller')
fix_files('internal/web')
