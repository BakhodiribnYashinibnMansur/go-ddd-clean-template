import os

# Define the directory to search in
search_dirs = [
    '/Users/mrb/Desktop/GCA/internal/controller',
    '/Users/mrb/Desktop/GCA/internal/web'
]

# Simple string replacements to avoid regex escaping issues
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

def fix_files():
    for directory in search_dirs:
        print(f"Checking directory: {directory}")
        for root, dirs, files in os.walk(directory):
            for file in files:
                if file.endswith('.go'):
                    filepath = os.path.join(root, file)
                    try:
                        with open(filepath, 'r') as f:
                            content = f.read()
                        
                        new_content = content
                        for old, new in replacements:
                            new_content = new_content.replace(old, new)
                        
                        if new_content != content:
                            with open(filepath, 'w') as f:
                                f.write(new_content)
                            print(f"Fixed: {filepath}")
                    except Exception as e:
                        print(f"Error processing {filepath}: {e}")

if __name__ == "__main__":
    fix_files()
