package encryption_test

import (
	"context"
	"github.com/riotkit-org/volume-syncing-controller/pkg/apis/riotkit.org/v1alpha1"
	appContext "github.com/riotkit-org/volume-syncing-controller/pkg/server/context"
	"github.com/riotkit-org/volume-syncing-controller/pkg/server/encryption"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1Fake "k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestAttachAutomaticEncryption_IsDisabled(t *testing.T) {
	definition := v1alpha1.PodFilesystemSync{}
	definition.ObjectMeta.Name = "definition-a"
	definition.ObjectMeta.Namespace = "default"

	// this is set to false, so no action should be taken
	definition.Spec.AutomaticEncryption.Enabled = false

	params := appContext.SynchronizationParameters{}
	kubeClient := v1Fake.NewSimpleClientset()

	assert.Nil(t, encryption.AttachAutomaticEncryption(&definition, &params, kubeClient))
}

func TestAttachAutomaticEncryption_MissingSecret(t *testing.T) {
	definition := v1alpha1.PodFilesystemSync{}
	definition.ObjectMeta.Name = "definition-a"
	definition.ObjectMeta.Namespace = "default"

	definition.Spec.AutomaticEncryption.Enabled = true
	definition.Spec.AutomaticEncryption.SecretName = ""

	params := appContext.SynchronizationParameters{}
	kubeClient := v1Fake.NewSimpleClientset()

	assert.Equal(t, encryption.AttachAutomaticEncryption(&definition, &params, kubeClient).Error(), "Missing `kind: Secret` reference in .spec.automaticEncryption.secretName")
}

func TestAttachAutomaticEncryption_GeneratesNewSecret(t *testing.T) {
	definition := v1alpha1.PodFilesystemSync{}
	definition.ObjectMeta.Name = "definition-a"
	definition.ObjectMeta.Namespace = "default"

	// WE ENABLE Automatic Encryption
	definition.Spec.AutomaticEncryption.Enabled = true

	// AND we set a valid secret name
	definition.Spec.AutomaticEncryption.SecretName = "pssst-dont-wait-for-salvation-start-revolution"

	params := appContext.SynchronizationParameters{RemotePath: "/bakunin-was-right-over-marx", Env: make(map[string]string)}
	kubeClient := v1Fake.NewSimpleClientset()
	assert.Nil(t, encryption.AttachAutomaticEncryption(&definition, &params, kubeClient))

	// check that our method stored a secret in API
	createdSecret, err := kubeClient.CoreV1().Secrets("default").Get(context.TODO(), "pssst-dont-wait-for-salvation-start-revolution", v1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, "pssst-dont-wait-for-salvation-start-revolution", createdSecret.Name, "The API should return a secret with desired name")

	_, keyExists := createdSecret.StringData["ENCRYPTED_PASSWORD"]
	assert.True(t, keyExists, "Expected that the .stringData of `kind: Secret` will contain ENCRYPTED_PASSWORD key")
}

func TestAttachAutomaticEncryption_UpdatesExistingSecretWhenItDoesNotContainValidKey(t *testing.T) {
	definition := v1alpha1.PodFilesystemSync{}
	definition.ObjectMeta.Name = "key-test"
	definition.ObjectMeta.Namespace = "git-clone-controller"

	// WE ENABLE Automatic Encryption
	definition.Spec.AutomaticEncryption.Enabled = true

	// AND we set a valid secret name
	definition.Spec.AutomaticEncryption.SecretName = "overthrow-gov-and-capital"

	// AND we create a secret before testing our AttachAutomaticEncryption() method
	existingSecret := &v12.Secret{}
	existingSecret.Name = "overthrow-gov-and-capital"
	existingSecret.Namespace = "git-clone-controller"
	existingSecret.StringData = map[string]string{
		"GOV": "is-a-sucker",
	}

	params := appContext.SynchronizationParameters{RemotePath: "/bakunin-was-right-over-marx", Env: make(map[string]string)}
	kubeClient := v1Fake.NewSimpleClientset(existingSecret) // AND API would already have this secret before we will start testing our method
	assert.Nil(t, encryption.AttachAutomaticEncryption(&definition, &params, kubeClient))

	// THEN expect that the secret will contain two keys instead of one
	modifiedSecret, err := kubeClient.CoreV1().Secrets("git-clone-controller").Get(context.TODO(), "overthrow-gov-and-capital", v1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, "is-a-sucker", modifiedSecret.StringData["GOV"])
	assert.NotEmptyf(t, modifiedSecret.StringData["ENCRYPTED_PASSWORD"], "Expected that ENCRYPTED_PASSWORD will be present and not empty")
}
