package encryption

import (
	"context"
	"crypto/rand"
	"github.com/pkg/errors"
	"github.com/riotkit-org/volume-syncing-operator/pkg/apis/riotkit.org/v1alpha1"
	appContext "github.com/riotkit-org/volume-syncing-operator/pkg/server/context"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"math/big"
)

// AttachAutomaticEncryption is mutating SynchronizationParameters by adding an encryption secret and setting up parameters
//                           the encryption secret is generated and placed in `kind: Secret` if it was not created previously
//                           So, an existing secret could be used, or a new one could be generated
func AttachAutomaticEncryption(syncDefinition *v1alpha1.PodFilesystemSync, params *appContext.SynchronizationParameters, kubeClient kubernetes.Interface) error {
	if !syncDefinition.Spec.AutomaticEncryption.Enabled {
		logrus.Debugf("Automatic encryption for %v is disabled", syncDefinition.TypeMeta.String())
		return nil
	}
	if syncDefinition.Spec.AutomaticEncryption.SecretName == "" {
		return errors.New("Missing `kind: Secret` reference in .spec.automaticEncryption.secretName")
	}

	logrus.Debug("Using encryption")

	// will generate a new secret only if secret with the same name does not exist
	// if it does not contain required key in .data, then the secret will be updated
	if err := generateSecretIfNotExists(kubeClient, syncDefinition.Namespace, syncDefinition.Spec.AutomaticEncryption.SecretName); err != nil {
		return errors.Wrap(err, "Cannot create secret for automatic encryption")
	}

	params.EnvSecrets = append(params.EnvSecrets, syncDefinition.Spec.AutomaticEncryption.SecretName)
	params.Env["ENCRYPTED_REMOTE"] = "remote:/" + params.RemotePath // there we lock encryption remote wrapper to original remote under selected directory
	params.Env["ENCRYPTED_FILENAME_ENCRYPTION"] = "obfuscate"

	return nil
}

// generateSecretIfNotExists will attempt to create or update the secret in Kubernetes
func generateSecretIfNotExists(kubeClient kubernetes.Interface, ns string, name string) error {
	secretApi := kubeClient.CoreV1().Secrets(ns)
	secret, _ := secretApi.Get(context.TODO(), name, v1.GetOptions{})

	// existing secret
	if secret.Name == name {
		if _, exists := secret.Data[appContext.SecretKeyNameForEncryption]; !exists {
			generatedPassword, genErr := generateRclonePassword(64, 128)
			if genErr != nil {
				return errors.Wrap(genErr, "The secret does not exist, or is invalid. Attempted to generate a new one, but got an error. Maybe not enough entropy?")
			}
			secret.StringData[appContext.SecretKeyNameForEncryption] = generatedPassword
		}

		if _, err := secretApi.Update(context.TODO(), secret, v1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "Cannot update secret for automatic encryption")
		}
	} else {
		// new secret
		generatedPassword, genErr := generateRclonePassword(64, 128)
		if genErr != nil {
			return errors.Wrap(genErr, "The secret does not exist, or is invalid. Attempted to generate a new one, but got an error. Maybe not enough entropy?")
		}

		secret.Name = name
		secret.Namespace = ns
		secret.StringData = map[string]string{
			appContext.SecretKeyNameForEncryption: generatedPassword,
		}

		if _, err := secretApi.Create(context.TODO(), secret, v1.CreateOptions{}); err != nil {
			return errors.Wrap(err, "Cannot create secret for automatic encryption")
		}
		return nil
	}

	return nil
}

// generateRclonePassword generates a string of random length, min n, max n+nPlusMax characters and obscures it with rclone-compatible method
func generateRclonePassword(n int, nPlusMax int) (string, error) {
	nPlus, _ := rand.Int(rand.Reader, big.NewInt(int64(nPlusMax)))
	n = n + int(nPlus.Uint64())

	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_./"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	obscured, err := Obscure(string(ret))
	if err != nil {
		return "", errors.Wrap(err, "Cannot obscure password (rclone method)")
	}

	return obscured, nil
}
