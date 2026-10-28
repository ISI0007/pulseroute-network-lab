[CmdletBinding()]
param(
    [Parameter(Mandatory=$true)][string]$Executable,
    [Parameter(Mandatory=$true)][string]$EvidenceDirectory
)
# CI-only native UI Automation. No remote-debugging endpoint or security overrides.
$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
if($env:GITHUB_ACTIONS -ne 'true'){throw 'This harness is CI-only. Do not use it to retry locally blocked application launches.'}
Add-Type -AssemblyName UIAutomationClient
Add-Type -AssemblyName UIAutomationTypes
$evidence=[IO.Path]::GetFullPath($EvidenceDirectory)
New-Item -ItemType Directory -Force $evidence | Out-Null
$data=Join-Path $evidence 'isolated-app-data'
$checks=[Collections.Generic.List[string]]::new()
$status='FAIL'; $reason=''; $appProcess=$null; $script:root=$null
$oldData=$env:PULSEROUTE_DATA
function Find-Button([string]$name) {
    $condition=[System.Windows.Automation.AndCondition]::new(
        [System.Windows.Automation.PropertyCondition]::new([System.Windows.Automation.AutomationElement]::NameProperty,$name),
        [System.Windows.Automation.PropertyCondition]::new([System.Windows.Automation.AutomationElement]::ControlTypeProperty,[System.Windows.Automation.ControlType]::Button))
    return $script:root.FindFirst([System.Windows.Automation.TreeScope]::Descendants,$condition)
}
function Wait-Button([string]$name) {
    $deadline=[DateTime]::UtcNow.AddSeconds(20)
    do {
        $button=Find-Button $name
        if($null -ne $button){return $button}
        if($appProcess.HasExited){throw 'Application exited during UI interaction'}
        Start-Sleep -Milliseconds 250
    } while([DateTime]::UtcNow -lt $deadline)
    throw "Native UI button not found: $name"
}
function Invoke-Button([string]$name) {
    $button=Wait-Button $name
    $pattern=$button.GetCurrentPattern([System.Windows.Automation.InvokePattern]::Pattern)
    $pattern.Invoke()
}
function Read-Exactly($stream,[int]$count) {
    $bytes=[byte[]]::new($count);$offset=0
    while($offset -lt $count){$n=$stream.Read($bytes,$offset,$count-$offset);if($n -eq 0){throw 'Unexpected socket EOF'};$offset+=$n}
    return ,$bytes
}
try {
    $runtime=Get-ChildItem -Path "${env:ProgramFiles(x86)}\Microsoft\EdgeWebView\Application\*\msedgewebview2.exe","$env:ProgramFiles\Microsoft\EdgeWebView\Application\*\msedgewebview2.exe" -ErrorAction SilentlyContinue | Select-Object -First 1
    if($null -eq $runtime){$status='BLOCKED';throw 'WebView2 runtime is not installed on this CI runner; no security settings or runtime installations were changed'}
    $env:PULSEROUTE_DATA=$data
    $appProcess=Start-Process -FilePath ([IO.Path]::GetFullPath($Executable)) -WorkingDirectory ([IO.Path]::GetDirectoryName([IO.Path]::GetFullPath($Executable))) -WindowStyle Normal -PassThru -RedirectStandardOutput (Join-Path $evidence 'native-stdout.txt') -RedirectStandardError (Join-Path $evidence 'native-stderr.txt')
    $deadline=[DateTime]::UtcNow.AddSeconds(45)
    do {
        $appProcess.Refresh()
        if($appProcess.HasExited){throw "Native app exited before window readiness, code $($appProcess.ExitCode)"}
        if($appProcess.MainWindowHandle -ne 0){break}
        Start-Sleep -Milliseconds 250
    } while([DateTime]::UtcNow -lt $deadline)
    if($appProcess.MainWindowHandle -eq 0){$status='BLOCKED';throw 'No accessible native window in the hosted runner desktop session'}
    $script:root=[System.Windows.Automation.AutomationElement]::FromHandle($appProcess.MainWindowHandle)
    $checks.Add('Native PulseRoute.exe window created with real WebView2')
    $elements=$script:root.FindAll([System.Windows.Automation.TreeScope]::Descendants,[System.Windows.Automation.Condition]::TrueCondition)
    $tree=foreach($element in $elements){[ordered]@{name=$element.Current.Name;type=$element.Current.ControlType.ProgrammaticName}}
    $tree | Select-Object -First 250 | ConvertTo-Json -Depth 3 | Set-Content -Encoding UTF8 (Join-Path $evidence 'native-accessibility-tree.json')
    foreach($page in @('Overview','Messages','Transfers','Network Lab','History','Settings')){
        Invoke-Button $page
        Start-Sleep -Milliseconds 400
        $checks.Add("Native navigation invoked: $page")
    }
    try {
        Add-Type -AssemblyName System.Drawing
        $rect=$script:root.Current.BoundingRectangle
        $bitmap=[Drawing.Bitmap]::new([int]$rect.Width,[int]$rect.Height)
        $graphics=[Drawing.Graphics]::FromImage($bitmap)
        try {$graphics.CopyFromScreen([int]$rect.X,[int]$rect.Y,0,0,$bitmap.Size);$bitmap.Save((Join-Path $evidence 'native-settings.png'),[Drawing.Imaging.ImageFormat]::Png)} finally {$graphics.Dispose();$bitmap.Dispose()}
    } catch {Write-Output 'Native screenshot unavailable; UI Automation/bridge checks continue'}
    Invoke-Button 'Start receiver'
    $null=Wait-Button 'Stop receiver'
    # Verify the actual listener started by the native UI using a v1 TCP ping.
    $client=[Net.Sockets.TcpClient]::new()
    try {
        $client.Connect('127.0.0.1',42191);$stream=$client.GetStream();$stream.ReadTimeout=5000;$stream.WriteTimeout=5000
        $body=[Text.Encoding]::UTF8.GetBytes('{"version":1,"op":"ping","id":"00112233445566778899aabbccddeeff"}')
        $prefix=[BitConverter]::GetBytes([uint32]($body.Length+1));[Array]::Reverse($prefix)
        $stream.Write($prefix,0,4);$stream.WriteByte(1);$stream.Write($body,0,$body.Length)
        $prefix=Read-Exactly $stream 4;[Array]::Reverse($prefix);$length=[BitConverter]::ToUInt32($prefix,0)
        if($length -lt 2 -or $length -gt 131072){throw 'Invalid native listener response length'}
        $frame=Read-Exactly $stream $length
        if($frame[0] -ne 1){throw 'Expected native listener control response'}
        $reply=[Text.Encoding]::UTF8.GetString($frame,1,$frame.Length-1) | ConvertFrom-Json
        if($reply.op -ne 'pong' -or $reply.id -ne '00112233445566778899aabbccddeeff'){throw 'Native listener returned wrong ping response'}
    } finally {$client.Dispose()}
    $checks.Add('UI receiver action started real TCP listener; v1 ping/pong passed')
    Invoke-Button 'Use Ocean theme'
    Start-Sleep -Milliseconds 600
    $checks.Add('Ocean palette selected through native UI')
    Invoke-Button 'Save node settings'
    Start-Sleep -Milliseconds 500
    $checks.Add('Save node settings invoked through native UI')
    Invoke-Button 'Stop receiver'
    $null=Wait-Button 'Start receiver'
    $checks.Add('Receiver stopped through native UI')
    $null=$appProcess.CloseMainWindow()
    if(-not $appProcess.WaitForExit(8000)){throw 'Native window did not close gracefully'}
    $database=Join-Path $data 'history.db'
    if(-not (Test-Path -LiteralPath $database)){throw 'Native UI did not create its isolated SQLite database'}
    $settings = python -c "import sqlite3,sys; c=sqlite3.connect(sys.argv[1]); print(c.execute('SELECT value FROM settings WHERE key=?',('node',)).fetchone()[0]); c.close()" $database
    if($LASTEXITCODE){throw 'Native settings persistence query failed'}
    $saved=$settings | ConvertFrom-Json
    if($saved.tcp -ne '127.0.0.1:42191'){throw 'Native saved settings did not match UI values'}
    $savedTheme = python -c "import sqlite3,sys; c=sqlite3.connect(sys.argv[1]); print(c.execute('SELECT value FROM settings WHERE key=?',('theme',)).fetchone()[0]); c.close()" $database
    if($LASTEXITCODE -or $savedTheme -ne 'ocean'){throw 'Native theme selection did not persist in SQLite'}
    $checks.Add('Native UI settings and Ocean palette persisted in SQLite; graceful window shutdown')
    $status='PASS'
} catch {
    $reason=$_.Exception.Message
} finally {
    if($null -ne $appProcess -and -not $appProcess.HasExited){
        $null=$appProcess.CloseMainWindow()
        if(-not $appProcess.WaitForExit(8000)){Stop-Process -Id $appProcess.Id; if($status -eq 'PASS'){$status='FAIL';$reason='Native window did not close gracefully'}}
    }
    $env:PULSEROUTE_DATA=$oldData
    [ordered]@{status=$status;reason=$reason;launch='Visible native window on isolated CI runner only';scope='Actual Windows Wails/WebView2 window and UIA actions; loopback sockets; not two-device or complete dialog/drag-drop verification';checks=$checks.ToArray();recordedAtUTC=[DateTime]::UtcNow.ToString('o')} | ConvertTo-Json -Depth 5 | Set-Content -Encoding UTF8 (Join-Path $evidence 'native-ui-result.json')
}
Get-Content (Join-Path $evidence 'native-ui-result.json')
if($status -eq 'BLOCKED'){exit 2}
if($status -ne 'PASS'){exit 1}
