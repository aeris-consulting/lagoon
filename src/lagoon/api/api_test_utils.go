package api

import "lagoon/datasource"

func ClearDatasources() {
	dataSources = make(map[datasource.DataSourceId]datasource.DataSource)
	DataSourcesHeaders = make(map[datasource.DataSourceId]DataSourceHeader)
}
