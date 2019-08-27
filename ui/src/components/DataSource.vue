<template>
    <div>
        <div v-if="selectedDatasource === null">
            <v-btn @click="refresh()" color="primary">Refresh data sources</v-btn>

            <div class="datasources" v-if="datasources !== null && datasources.length > 0">
                <v-list>
                    <v-list-item
                        :key="datasource.uuid"
                        @click="select(datasource)"
                        v-for="datasource in datasources">
                        <v-list-item-content>
                            <v-list-item-title>{{ datasource.name }}</v-list-item-title>
                            <v-list-item-subtitle v-if="datasource.description">{{ datasource.description }}</v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>
                </v-list>
            </div>
        </div>

        <div v-else>
            <splitpanes style="height: 400px" vertical>
                <entrypoint-list @display-modal="showConfirmation" splitpanes-min=10
                                 v-bind:dataSource="selectedDatasource"></entrypoint-list>
                <entrypoint-content :key="n.getFullName()"
                                    @display-modal="showConfirmation"
                                    v-bind:dataSource="selectedDatasource" v-bind:node="n"
                                    v-for="n in selectedDatasource.selectedNodes"></entrypoint-content>
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

    .datasources {
        margin-top: 10px;
    }


</style>
