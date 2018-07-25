package main

import "os"

var inputs = struct {
sshRsaPrivateKey string
sshKeySavePath string
isRemoveOtherIdentities string
testInput string
verbose string
}{
sshRsaPrivateKey: os.Getenv("ssh_rsa_private_key"),
sshKeySavePath: os.Getenv("ssh_key_save_path"),
isRemoveOtherIdentities: os.Getenv("is_remove_other_identities"),
testInput: os.Getenv("test_input"),
verbose: os.Getenv("verbose"),
}