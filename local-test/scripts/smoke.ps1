param(
  [string]$HostAddr = '127.0.0.1',
  [int]$Port = 3333
)

$ErrorActionPreference = 'Stop'

$here = Split-Path -Parent $MyInvocation.MyCommand.Path
$script = Join-Path $here 'smoke_test_mcp_tcp.py'

python $script --host $HostAddr --port $Port
