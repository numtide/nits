Require nits

Set Shell zsh
Set FontSize 14

Set Width 1440
Set Height 800

Type "nits agent ls"
Enter
Sleep 500ms

Enter
Type "nits agent deploy --action test --name test-vm-1 .#nixosConfigurations.x86_64-linux_test-vm-1-no-hello.config.system.build.toplevel"
Sleep 500ms
Enter

Sleep 20s
