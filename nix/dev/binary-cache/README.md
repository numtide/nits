# Dev Binary Cache

## Details

| Property   | Value                                                     |
| ---------- | --------------------------------------------------------- |
| Url        | `http://localhost:5050`                                   |
| Url in VM  | `http://10.0.0.2:5050`                                    |
| Public key | `nits-dev-1:o8ObSMff6cFYpBHf5f2ghUEqDmjP8opj1sA6w4eIys4=` |

## Generate a new key pair

```shell
nix key generate-secret --key-name nits-dev-1 > ./key.sec
nix key convert-secret-to-public < ./key.sec > ./key.pub
```
