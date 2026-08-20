import os
import glob

api_dir = 'backend/api'
serverless_dir = 'backend/pkg/serverless'
os.makedirs(serverless_dir, exist_ok=True)

files = glob.glob(f'{api_dir}/**/*.go', recursive=True)

for f in files:
    if not os.path.isfile(f):
        continue
    
    # Generate a unique handler name
    # e.g., backend/api/auth/login/login.go -> AuthLogin
    parts = f.replace(api_dir + '/', '').replace('.go', '').split('/')
    
    # Filter out index and duplicate parts
    name_parts = []
    for p in parts:
        if p != 'index' and p not in name_parts:
            name_parts.append(p.capitalize())
        elif p == 'index' and len(parts) == 2:
            # e.g., groups/index/index.go -> GroupsIndex
            pass
            
    if parts[-1] == 'index':
        # groups/index/index.go -> Groups
        # groups/expenses/index/index.go -> GroupsExpenses
        func_prefix = "".join([p.capitalize() for p in parts[:-2]]) + parts[-2].capitalize()
    else:
        # auth/login/login.go -> AuthLogin
        func_prefix = "".join([p.capitalize() for p in parts[:-1]])

    if not func_prefix:
        func_prefix = "Root"
    
    # Read file
    with open(f, 'r') as file:
        content = file.read()
        
    # Replace package handler with package serverless
    content = content.replace('package handler', 'package serverless')
    
    # Replace func Handler with func Handle{Prefix}
    content = content.replace('func Handler(', f'func Handle{func_prefix}(')
    
    # Handle the comment
    content = content.replace('// Handler xử lý', f'// Handle{func_prefix} xử lý')

    # Save to serverless_dir
    dest_file = os.path.join(serverless_dir, f"{func_prefix.lower()}.go")
    with open(dest_file, 'w') as file:
        file.write(content)
        
    print(f"Migrated {f} to {dest_file} (Handle{func_prefix})")

