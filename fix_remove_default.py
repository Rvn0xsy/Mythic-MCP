with open('pkg/server/tools_payloads.go', 'r') as f:
    content = f.read()

# Find the defaultHTTPxConfig const and remove it entirely
marker = '// defaultHTTPxConfig is a minimal working httpx malleable C2 configuration'
idx = content.find(marker)
if idx >= 0:
    # Go back to find the preceding newline
    while idx > 0 and content[idx-1] == '\n':
        idx -= 1
    content = content[:idx] + '\n'
    print("Removed defaultHTTPxConfig const")
else:
    print("ERROR: could not find defaultHTTPxConfig")

with open('pkg/server/tools_payloads.go', 'w') as f:
    f.write(content)
