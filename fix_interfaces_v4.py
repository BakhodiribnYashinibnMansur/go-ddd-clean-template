import os

search_dirs = [
    '/Users/mrb/Desktop/GCA/internal/controller',
    '/Users/mrb/Desktop/GCA/internal/web',
    '/Users/mrb/Desktop/GCA/pkg/asynq'
]

replacements = [
    ('.User.Client.', '.User.Client().'),
    ('.User.Session.', '.User.Session().'),
    ('.Audit.Log.', '.Audit.Log().'),
    ('.Audit.LogUC().', '.Audit.Log().'), # Specifically for LogUC stragglers
    ('.Audit.History.', '.Audit.History().'),
    ('.Audit.Metric.', '.Audit.Metric().'),
    ('.Audit.SystemError.', '.Audit.SystemError().'),
    ('.Authz.Role.', '.Authz.Role().'),
    ('.Authz.Permission.', '.Authz.Permission().'),
    ('.Authz.Policy.', '.Authz.Policy().'),
    ('.Authz.Relation.', '.Authz.Relation().'),
    ('.Authz.Scope.', '.Authz.Scope().'),
    ('.SiteSetting.', '.SiteSetting().'), # In case SiteSetting is also an interface now
]

processed_count = 0
fixed_count = 0

for directory in search_dirs:
    print(f"Checking directory: {directory}")
    for root, dirs, files in os.walk(directory):
        for file in files:
            if file.endswith('.go'):
                filepath = os.path.join(root, file)
                processed_count += 1
                try:
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
                except Exception as e:
                    print(f"Error processing {filepath}: {e}")

print(f"Summary: Processed {processed_count} files, fixed {fixed_count} files.")
