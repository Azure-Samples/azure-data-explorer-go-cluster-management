package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kusto/armkusto"
	"github.com/olekukonko/tablewriter"
)

const (
	subscriptionEnvVar      = "AZURE_SUBSCRIPTION_ID"
	resourceGroupEnvVar     = "AZURE_RESOURCE_GROUP"
	locationEnvVar          = "AZURE_LOCATION"
	clusterNamePrefixEnvVar = "CLUSTER_NAME_PREFIX"
	dbNamePrefixEnvVar      = "DATABASE_NAME_PREFIX"

	clusterName  = "ADXTestCluster"
	databaseName = "ADXTestDB"
)

var (
	subscription      string
	rgName            string
	location          string
	clusterNamePrefix string
	dbNamePrefix      string
)

func init() {
	subscription = os.Getenv(subscriptionEnvVar)
	if subscription == "" {
		log.Fatalf("missing environment variable %s", subscriptionEnvVar)
	}

	rgName = os.Getenv(resourceGroupEnvVar)
	if rgName == "" {
		log.Fatalf("missing environment variable %s", resourceGroupEnvVar)
	}

	location = os.Getenv(locationEnvVar)
	if location == "" {
		log.Fatalf("missing environment variable %s", locationEnvVar)
	}

	clusterNamePrefix = os.Getenv(clusterNamePrefixEnvVar)
	if clusterNamePrefix == "" {
		log.Fatalf("missing environment variable %s", clusterNamePrefixEnvVar)
	}

	dbNamePrefix = os.Getenv(dbNamePrefixEnvVar)
	if dbNamePrefix == "" {
		log.Fatalf("missing environment variable %s", dbNamePrefixEnvVar)
	}
}

func main() {
	createCluster(subscription, clusterNamePrefix+clusterName, location, rgName)
	listClusters(subscription, rgName)
	createDatabase(subscription, rgName, clusterNamePrefix+clusterName, location, dbNamePrefix+databaseName)
	listDatabases(subscription, rgName, clusterNamePrefix+clusterName)
	deleteDatabase(subscription, rgName, clusterNamePrefix+clusterName, dbNamePrefix+databaseName)
	deleteCluster(subscription, clusterNamePrefix+clusterName, rgName)
}

func getClustersClient(subscription string) *armkusto.ClustersClient {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	client, err := armkusto.NewClustersClient(subscription, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

// 1 instance, Basic tier with compute type Dev(No SLA)_Standard_D11_v2
func createCluster(sub, name, location, rgName string) {
	ctx := context.Background()

	numInstances := int32(1)
	client := getClustersClient(sub)
	result, err := client.BeginCreateOrUpdate(
		ctx,
		rgName,
		name,
		armkusto.Cluster{
			Location: &location,
			SKU: &armkusto.AzureSKU{
				Name:     to.Ptr(armkusto.AzureSKUNameDevNoSLAStandardD11V2),
				Capacity: &numInstances,
				Tier:     to.Ptr(armkusto.AzureSKUTierBasic),
			},
		},
		nil,
	)
	if err != nil {
		log.Fatal("failed to start cluster creation ", err)
	}

	log.Printf("waiting for cluster creation to complete - %s\n", name)
	r, err := result.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("created cluster %s\n", *r.Name)
}

func listClusters(sub, rgName string) {
	log.Printf("listing clusters in resource group %s\n", rgName)
	ctx := context.Background()

	result := getClustersClient(sub).NewListByResourceGroupPager(rgName, nil)

	data := [][]string{}

	for result.More() {
		temp, err := result.NextPage(ctx)
		if err != nil {
			log.Fatal(err)
		}
		for _, c := range temp.Value {
			data = append(data, []string{*c.Name, string(*c.Properties.State), *c.Location, strconv.Itoa(int(*c.SKU.Capacity)), *c.Properties.URI})
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "State", "Location", "Instances", "URI"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

func getDBClient(subscription string) *armkusto.DatabasesClient {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	client, err := armkusto.NewDatabasesClient(subscription, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func createDatabase(sub, rgName, clusterName, location, dbName string) {
	ctx := context.Background()

	client := getDBClient(sub)
	future, err := client.BeginCreateOrUpdate(ctx, rgName, clusterName, dbName, &armkusto.ReadWriteDatabase{Kind: to.Ptr(armkusto.KindReadWrite), Location: &location}, nil)

	if err != nil {
		log.Fatal("failed to start database creation ", err)
	}

	log.Printf("waiting for database creation to complete - %s\n", dbName)
	resp, err := future.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	kdb := resp.GetDatabase()
	log.Printf("created DB %s with ID %s and type %s\n", *kdb.Name, *kdb.ID, *kdb.Type)
}

func listDatabases(sub, rgName, clusterName string) {
	log.Printf("listing databases in cluster %s\n", clusterName)

	ctx := context.Background()
	result := getDBClient(sub).NewListByClusterPager(rgName, clusterName, nil)

	data := [][]string{}

	for result.More() {
		temp, err := result.NextPage(ctx)
		if err != nil {
			log.Fatal(err)
		}
		for _, db := range temp.Value {
			if *db.GetDatabase().Kind == armkusto.KindReadWrite {
				data = append(data, []string{*db.GetDatabase().Name, string(*db.GetDatabase().Kind), *db.GetDatabase().Location, *db.GetDatabase().Type})
			}
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "State", "Location", "Type"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

func deleteDatabase(sub, rgName, clusterName, dbName string) {
	ctx := context.Background()

	future, err := getDBClient(sub).BeginDelete(ctx, rgName, clusterName, dbName, nil)
	if err != nil {
		log.Fatal("failed to start database deletion ", err)
	}

	log.Printf("waiting for database deletion to complete - %s\n", dbName)

	_, err = future.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatal("database deletion process has not yet completed", err)
	}

	log.Printf("deleted DB %s from cluster %s", dbName, clusterName)
}

func deleteCluster(sub, clusterName, rgName string) {
	ctx := context.Background()

	client := getClustersClient(sub)
	result, err := client.BeginDelete(ctx, rgName, clusterName, nil)

	if err != nil {
		log.Fatal("failed to start cluster deletion ", err)
	}

	log.Printf("waiting for cluster deletion to complete - %s\n", clusterName)

	_, err = result.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatal("cluster deletion failed ", err)
	}

	log.Printf("deleted ADX cluster %s from resource group %s", clusterName, rgName)
}
