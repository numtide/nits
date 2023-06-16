Require deploy-agent

Set FontSize 14

Set Width 1024
Set Height 800

Type "# How to deploy an agent"
Sleep 500ms
Enter
Enter

Type "deploy-agent 1 test '{ config.systemd.services.hello.enable = false; }'"
Sleep 500ms
Enter

Sleep 30s
Ctrl+c

Sleep 5s
