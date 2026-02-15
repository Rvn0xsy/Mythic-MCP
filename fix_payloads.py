import re

with open('pkg/server/tools_payloads.go', 'r') as f:
    content = f.read()

# Fix the resolveFileParams call: params, err = ... -> resolved, resolveErr := ...
# and add params = resolved after the if block
old_block = '''\t\t\tif params != nil {
\t\t\t\tparams, err = s.resolveFileParams(ctx, name, params)
\t\t\t\tif err != nil {
\t\t\t\t\treturn nil, nil, fmt.Errorf("failed to resolve file params for C2 profile %q: %w", name, err)
\t\t\t\t}
\t\t\t}'''

new_block = '''\t\t\tif params != nil {
\t\t\t\tresolved, resolveErr := s.resolveFileParams(ctx, name, params)
\t\t\t\tif resolveErr != nil {
\t\t\t\t\treturn nil, nil, fmt.Errorf("failed to resolve file params for C2 profile %q: %w", name, resolveErr)
\t\t\t\t}
\t\t\t\tparams = resolved
\t\t\t}'''

if old_block in content:
    content = content.replace(old_block, new_block)
    print("Fixed resolveFileParams block")
else:
    print("Could not find old block, trying alternate whitespace...")
    # Try to find it with a regex
    pattern = r'if params != nil \{\s*params, err = s\.resolveFileParams\(ctx, name, params\)\s*if err != nil \{\s*return nil, nil, fmt\.Errorf\("failed to resolve file params for C2 profile %q: %w", name, err\)\s*\}\s*\}'
    match = re.search(pattern, content)
    if match:
        content = content[:match.start()] + new_block + content[match.end():]
        print("Fixed with regex")
    else:
        print("ERROR: Could not find the block to fix")

with open('pkg/server/tools_payloads.go', 'w') as f:
    f.write(content)
