package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/kusto/mgmt/kusto"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/olekukonko/tablewriter"
)

const (
	subscriptionEnvVar      = "SUBSCRIPTION"
	resourceGroupEnvVar     = "RESOURCE_GROUP"
	locationEnvVar          = "LOCATION"
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
		log.Fatalf("missing environment variable %s", rgName)
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

func getClustersClient(subscription string) kusto.ClustersClient {
	client := kusto.NewClustersClient(subscription)
	authR, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	client.Authorizer = authR

	return client
}

// 1 instance, Basic tier with compute type Dev(No SLA)_Standard_D11_v2
func createCluster(sub, name, location, rgName string) {
	ctx := context.Background()

	numInstances := int32(1)
	client := getClustersClient(sub)
	result, err := client.CreateOrUpdate(ctx, rgName, name, kusto.Cluster{Location: &location, Sku: &kusto.AzureSku{Name: kusto.DevNoSLAStandardD11V2, Capacity: &numInstances, Tier: kusto.Basic}})

	if err != nil {
		log.Fatal("failed to start cluster creation ", err)
	}

	log.Printf("waiting for cluster creation to complete - %s\n", name)
	err = result.WaitForCompletionRef(context.Background(), client.Client)
	if err != nil {
		log.Fatal("error during cluster creation ", err)
	}

	r, err := result.Result(client)
	if err != nil {
		log.Fatal("cluster creation failed ", err)
	}

	log.Printf("created cluster %s\n", *r.Name)
}

func listClusters(sub, rgName string) {
	log.Printf("listing clusters in resource group %s\n", rgName)
	ctx := context.Background()

	result, err := getClustersClient(sub).ListByResourceGroup(ctx, rgName)
	if err != nil {
		log.Fatal("failed to get clusters ", err)
	}

	data := [][]string{}

	for _, c := range *result.Value {
		data = append(data, []string{*c.Name, string(c.State), *c.Location, strconv.Itoa(int(*c.Sku.Capacity)), *c.URI})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "State", "Location", "Instances", "URI"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

func getDBClient(subscription string) kusto.DatabasesClient {
	client := kusto.NewDatabasesClient(subscription)
	authR, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	client.Authorizer = authR

	return client
}

func createDatabase(sub, rgName, clusterName, location, dbName string) {
	ctx := context.Background()

	client := getDBClient(sub)
	future, err := client.CreateOrUpdate(ctx, rgName, clusterName, dbName, kusto.ReadWriteDatabase{Kind: kusto.KindReadWrite, Location: &location})

	if err != nil {
		log.Fatal("failed to start database creation ", err)
	}

	log.Printf("waiting for database creation to complete - %s\n", dbName)
	err = future.WaitForCompletionRef(context.Background(), client.Client)
	if err != nil {
		log.Fatal("failed to create database ", err)
	}

	r, err := future.Result(client)
	if err != nil {
		log.Fatal("database creation failed ", err)
	}
	kdb, _ := r.Value.AsReadWriteDatabase()
	log.Printf("created DB %s with ID %s and type %s\n", *kdb.Name, *kdb.ID, *kdb.Type)
}

func listDatabases(sub, rgName, clusterName string) {
	log.Printf("listing databases in cluster %s\n", clusterName)

	ctx := context.Background()
	result, err := getDBClient(sub).ListByCluster(ctx, rgName, clusterName)
	if err != nil {
		log.Fatal("failed to get databases in cluster ", err)
	}

	data := [][]string{}

	for _, db := range *result.Value {
		rwDB, isRW := db.AsReadWriteDatabase()
		if isRW {
			data = append(data, []string{*rwDB.Name, string(rwDB.ProvisioningState), *rwDB.Location, *rwDB.Type})
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

	client := getDBClient(sub)
	future, err := getDBClient(sub).Delete(ctx, rgName, clusterName, dbName)

	if err != nil {
		log.Fatal("failed to start database deletion ", err)
	}

	log.Printf("waiting for database deletion to complete - %s\n", dbName)
	err = future.WaitForCompletionRef(context.Background(), client.Client)
	if err != nil {
		log.Fatal("failed to delete database ", err)
	}

	r, err := future.Result(client)
	if err != nil {
		log.Fatal("database deletion process has not yet completed", err)
	}

	if r.StatusCode == 200 {
		log.Printf("deleted DB %s from cluster %s", dbName, clusterName)
	} else {
		log.Println("failed to delete DB. response status code - ", r.StatusCode)
	}
}

func deleteCluster(sub, clusterName, rgName string) {
	ctx := context.Background()

	client := getClustersClient(sub)
	result, err := client.Delete(ctx, rgName, clusterName)

	if err != nil {
		log.Fatal("failed to start cluster deletion ", err)
	}

	log.Printf("waiting for cluster deletion to complete - %s\n", clusterName)
	err = result.WaitForCompletionRef(context.Background(), client.Client)
	if err != nil {
		log.Fatal("error during cluster deletion ", err)
	}

	r, err := result.Result(client)
	if err != nil {
		log.Fatal("cluster deletion failed ", err)
	}

	if r.StatusCode == 200 {
		log.Printf("deleted ADX cluster %s from resource group %s", clusterName, rgName)
	} else {
		log.Println("failed to delete ADX cluster. response status code - ", r.StatusCode)
	}
}
