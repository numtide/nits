Require nits

Set Shell zsh
Set FontSize 14

Set Width 1440
Set Height 800

Type "nits agent ls"
Enter
Sleep 500ms

Enter
Type "nits agent deploy --name test-vm-1 .#nixosConfigurations.x86_64-linux_test-vm-1-no-hello.config.system.build.toplevel"
Sleep 500ms
Enter

Sleep 10s

Type "nits agent logs --since 10m test-vm-1"
Sleep 500ms
Enter

Sleep 5s
