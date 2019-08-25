<template>
    <div id="datasources">
        <div class="container" v-if="selectedDatasource === null">
            <button @click="refresh()" class="btn btn-outline-primary">Refresh data sources</button>

            <div v-if="datasources !== null && datasources.length > 0">
                <div class="row header">
                    <div class="col-sm">
                        Name
                    </div>
                    <div class="col-sm">
                        Description
                    </div>
                </div>
                <div :key="datasource.uuid" @click="select(datasource)"
                     class="row item" v-for="datasource in datasources">
                    <div class="col-sm">
                        {{ datasource.name }}
                    </div>
                    <div class="col-sm">
                        <span class="description" v-if="datasource.description">{{ datasource.description }}</span><span
                            class="no-description" v-else>None</span>
                    </div>
                </div>
            </div>
        </div>

        <div class="container" v-else>
            <splitpanes style="height:available" vertical>
                <entrypoint-list @display-modal="showConfirmation" splitpanes-min=10
                                 v-bind:dataSource="selectedDatasource"></entrypoint-list>
                <entrypoint-content :key="n.getFullName()"
                                    @display-modal="showConfirmation"
                                    v-bind:dataSource="selectedDatasource" v-bind:node="n"
                                    v-for="n,i in selectedDatasource.selectedNodes"></entrypoint-content>
            </splitpanes>
        </div>
    </div>
</template>

<script>
    import axios from 'axios';
    import EntrypointList from "./EntrypointList";
    import EntrypointContent from "./EntrypointContent";
    import DataSource from "../models/DataSource";
    import Splitpanes from 'splitpanes'

    export default {
        name: 'DataSourceList',
        components: {EntrypointList, EntrypointContent, Splitpanes},

        data() {
            return {
                datasources: [],
                errors: [],
                selectedDatasource: null,
            }
        },

        methods: {
            showConfirmation: function (event) {
                this.$emit('display-modal', event);
            },

            refresh: function () {
                let root = '..';
                if (process.env.VUE_APP_API_SCHEME && process.env.VUE_APP_API_URL) {
                    root = process.env.VUE_APP_API_SCHEME + '://' + process.env.VUE_APP_API_URL;
                }
                axios.get(root + '/datasource')
                    .then(response => {
                        this.datasources = response.data.datasources;
                    })
                    .catch(e => {
                        this.errors.push(e)
                    });
            },

            select: function (datasource) {
                this.selectedDatasource = new DataSource(datasource.uuid, '');
            }
        },

        created() {
            this.refresh();
        }
    }
</script>

<style lang="scss" scoped>

    #datasources {

        .header {
            font-size: 1.1em;
            font-weight: bold;
        }

        .item {
            cursor: pointer;

            .no-description {
                font-style: italic;
            }

        }
    }


</style>
