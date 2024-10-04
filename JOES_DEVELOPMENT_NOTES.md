The following flags result in a successful run of the program, WITH source code links:

```bash
--dir
/Users/jburger/bin/AT_Eclipse
--accessToken ${LAUNCH_DARKLY_WRITE_SERVICE_TOKEN}
--projKey
eclipse
--repoName
AT_Eclipse
--defaultBranch
master
--repoUrl
git@ssh.dev.azure.com:v3/FirstAmCorp/Eclipse/AT_Eclipse
--commitUrlTemplate
"https://dev.azure.com/FirstAmCorp/Eclipse/_git/AT_Eclipse?version=GB${branchName}"
--hunkUrlTemplate
"https://dev.azure.com/FirstAmCorp/Eclipse/_git/AT_Eclipse?version=GBmaster&path=${filePath}"
```

Additionally, my ~/.ssh/config file has the following entry, but without the #'s commenting things out:

```bash
Host github.com
  AddKeysToAgent yes
  UseKeychain yes
  IdentityFile ~/.ssh/id_ed25519

#Host ssh.dev.azure.com
  #IdentityFile ~/.ssh/ado_personal_access_token_PAT
  #IdentitiesOnly yes
  #HostkeyAlgorithms +ssh-rsa
  #PubkeyAcceptedKeyTypes=ssh-rsa
~                                               
```