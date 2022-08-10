$scriptpath = $MyInvocation.MyCommand.Path
$dir = Split-Path $scriptpath
Push-Location $dir/..

Copy-Item -r ./dist hathoute-modbusrtu-datasource

$version = Read-Host -Prompt "Please provide the release version"

# Using wsl because grafana doesnt recognize archives made with Compress-Archive
wsl zip -r hathoute-modbus-datasource-$version.zip hathoute-modbus-datasource

Remove-Item -r hathoute-modbusrtu-datasource

Pop-Location