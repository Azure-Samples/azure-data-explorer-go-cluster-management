# Manage an Azure Data Explorer cluster & database using Azure Go SDK

In this example, you execute create, list, delete operations on Azure Data Explorer cluster and database using [Go](https://golang.org/). Azure Data Explorer is a fast, fully managed data analytics service for real-time analysis on large volumes of data streaming from applications, websites, IoT devices, and more. To use Azure Data Explorer, first create a cluster, and create one or more databases in that cluster. Then ingest, or load, data into a database so that you can run queries against it.

## Prerequisites

* If you don't have an Azure subscription, create a [free Azure account](https://azure.microsoft.com/free) before you begin.
* Install [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).
* Install an appropriate version of Go. The [Azure Go SDK](https://github.com/Azure/azure-sdk-for-go) officially supports the last two major releases of Go.

## Run the application

When you run the sample code as is, the following actions are performed:
    
- An Azure Data Explorer cluster is created.
- All the Azure Data Explorer clusters in the specified resource group are listed.
- An Azure Data Explorer database is created as a part of the cluster created earlier.
- All the databases in the specified cluster are listed.
- The database is deleted.
- The cluster is deleted.

> To try different combinations of operations, you can uncomment/comment the respective functions in `main.go`.

1. Clone the sample code from GitHub:

    ```console
    git clone https://github.com/Azure-Samples/azure-data-explorer-go-cluster-management.git
    cd azure-data-explorer-go-cluster-management
    ```

1. Run the sample code as seen in this snippet from `main.go`: 

    ```go
    func main() {
    	createCluster(subscription, clusterNamePrefix+clusterName, location, rgName)
    	listClusters(subscription, rgName)
    	createDatabase(subscription, rgName, clusterNamePrefix+clusterName, location, dbNamePrefix+databaseName)
    	listDatabases(subscription, rgName, clusterNamePrefix+clusterName)
    	deleteDatabase(subscription, rgName, clusterNamePrefix+clusterName, dbNamePrefix+databaseName)
    	deleteCluster(subscription, clusterNamePrefix+clusterName, rgName)
    }
    ```

1. Export required environment variables, including service principal information used to authenticate to Azure Data Explorer for executing cluster and operation operations. To create a service principal, use Azure CLI with the [az ad sp create-for-rbac](https://docs.microsoft.com/cli/azure/ad/sp?view=azure-cli-latest#az-ad-sp-create-for-rbac) command. Set the information with the cluster endpoint and the database name in the form of environment variables that will be used by the program:

    ```console
    export AZURE_CLIENT_ID="<enter service principal client ID>"
    export AZURE_CLIENT_SECRET="<enter service principal client secret>"
    export AZURE_TENANT_ID="<enter tenant ID>"
    
    export AZURE_SUBSCRIPTION_ID="<enter subscription ID>"
    export AZURE_RESOURCE_GROUP="<enter resource group name>"
    export AZURE_LOCATION="<enter azure location e.g. Southeast Asia>"

    export CLUSTER_NAME_PREFIX="<enter prefix. name of cluster [prefix]-ADXTestCluster>"
    export DATABASE_NAME_PREFIX="<enter prefix. name of database [prefix]-ADXTestDB>"
    ```

1. Run the program:

    ```console
    go run main.go
    ```

    You'll get a similar output:

    ```console
    waiting for cluster creation to complete - fooADXTestCluster
    created cluster fooADXTestCluster
    listing clusters in resource group <your resource group>
    +-------------------+---------+----------------+-----------+-----------------------------------------------------------+
    |       NAME        |  STATE  |    LOCATION    | INSTANCES |                            URI                           |
    +-------------------+---------+----------------+-----------+-----------------------------------------------------------+
    | fooADXTestCluster | Running | Southeast Asia |         1 | https://fooADXTestCluster.southeastasia.kusto.windows.net |
    +-------------------+---------+----------------+-----------+-----------------------------------------------------------+
    
    waiting for database creation to complete - barADXTestDB
    created DB fooADXTestCluster/barADXTestDB with ID /subscriptions/<your subscription ID>/resourceGroups/<your resource group>/providers/Microsoft.Kusto/Clusters/fooADXTestCluster/Databases/barADXTestDB and type Microsoft.Kusto/Clusters/Databases
    
    listing databases in cluster fooADXTestCluster
    +--------------------------------+-----------+----------------+------------------------------------+
    |              NAME              |   STATE   |    LOCATION    |                TYPE                |
    +--------------------------------+-----------+----------------+------------------------------------+
    | fooADXTestCluster/barADXTestDB | Succeeded | Southeast Asia | Microsoft.Kusto/Clusters/Databases |
    +--------------------------------+-----------+----------------+------------------------------------+
    
    waiting for database deletion to complete - barADXTestDB
    deleted DB barADXTestDB from cluster fooADXTestCluster

    waiting for cluster deletion to complete - fooADXTestCluster
    deleted ADX cluster fooADXTestCluster from resource group <your resource group>
    ```

## Clean up resources

If you did not delete the cluster programmatically using the sample code in this repo, you can do so manually using the Azure CLI

```azurecli
az kusto cluster delete --cluster-name <enter name> --resource-group <enter name>
```

## Resources

- [What is Azure Data Explorer?](https://docs.microsoft.com/en-us/azure/data-explorer/data-explorer-overview)
- [Create an Azure Data Explorer cluster and database using Azure Portal](https://docs.microsoft.com/en-us/azure/data-explorer/create-cluster-database-portal)
- [Ingest sample data into Azure Data Explorer](https://docs.microsoft.com/en-us/azure/data-explorer/ingest-sample-data)