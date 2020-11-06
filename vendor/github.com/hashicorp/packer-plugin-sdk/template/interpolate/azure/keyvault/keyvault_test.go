package keyvault

import "testing"

func TestParseAzureResourceIDValidCases(t *testing.T) {
	resourceID := "/subscriptions/mysubscription/resourceGroups/myresourcegroup/providers/Microsoft.KeyVault/vaults/myvault"

	rg, name, err := getKeyVaultResourceGroupAndName(resourceID)
	if err != nil {
		t.Fatalf("no error expected with %v but got %q", resourceID, err)
	}

	if rg != "myresourcegroup" {
		t.Fatalf("Expected Resource Group was myresourcegroup but found: %s", rg)
	}

	if name != "myvault" {
		t.Fatalf("Expected Resource Group was myvault but found: %s", name)
	}
}
