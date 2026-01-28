import os

search_dirs = [
    '/Users/mrb/Desktop/GCA/internal/controller',
    '/Users/mrb/Desktop/GCA/internal/web'
]

replacements = [
    ('.User.Client.', '.User.Client().'),
    ('.User.Session.', '.User.Session().'),
    ('.Audit.Log.', '.Audit.Log().'),
    ('.Audit.History.', '.Audit.History().'),
    ('.Audit.Metric.', '.Audit.Metric().'),
    ('.Audit.SystemError.', '.Audit.SystemError().'),
    ('.Authz.Role.', '.Authz.Role().'),
    ('.Authz.Permission.', '.Authz.Permission().'),
    ('.Authz.Policy.', '.Authz.Policy().'),
    ('.Authz.Relation.', '.Authz.Relation().'),
    ('.Authz.Scope.', '.Authz.Scope().'),
]

processed_count = 0
fixed_count = 0

for directory in search_dirs:
    for root, dirs, files in os.walk(directory):
        for file in files:
            if file.endswith('.go'):
                filepath = os.path.join(root, file)
                processed_count += 1
                with open(filepath, 'r') as f:
                    content = f.read()
                
                new_content = content
                for old, new in replacements:
                    new_content = new_content.replace(old, new)
                
                if new_content != content:
                    with open(filepath, 'w') as f:
                        f.write(new_content)
                    print(f"FIXED: {filepath}")
                    fixed_count += 1

print(f"Summary: Processed {processed_count} files, fixed {fixed_count} files.")
