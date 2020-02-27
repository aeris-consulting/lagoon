<template>
    <div>
        <div v-if="selectedDatasourceId === null">
            <v-btn @click="refresh()" color="primary">Refresh data sources</v-btn>

            <div class="datasources" v-if="datasources !== null && datasources.length > 0">
                <v-list>
                    <v-list-item
                            :key="datasource.id"
                            @click="select(datasource)"
                            v-for="datasource in datasources">
                        <v-list-item-content>
                            <v-list-item-title>{{ datasource.name }}
                                <template v-if="datasource.readonly">&nbsp;(Read-only)</template>
                            </v-list-item-title>
                            <v-list-item-subtitle v-if="datasource.description">{{ datasource.description }}</v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>
                </v-list>
            </div>
        </div>

        <div v-else>
            <splitpanes watchSlots vertical class="splitpanes">
                <div 
                    class="entrypoint-list-container"
                    splitpanes-min="20"
                    splitpanes-size="30">
                    <entrypoint-list
                                    @display-modal="showConfirmation"
                                    v-bind:datasourceId="selectedDatasourceId"></entrypoint-list>
                </div>
                <!-- <div class="details-container" 
                    splitpanes-size="70">
                    <template v-if="selectedDatasource.selectedNodes.length > 0">
                        <v-tabs
                            splitpanes-size="70"
                            background-color="primary"
                            dark>
                            <template v-for="n in selectedDatasource.selectedNodes">
                                <v-tab :key="n.getFullName()">
                                    <span class="tab-title" :title="n.getFullName()">
                                        {{ n.getFullName() }}
                                    </span>
                                </v-tab>
                                <v-tab-item :key="n.getFullName() + '-tab-item'">
                                    <entrypoint-content
                                                        @display-modal="showConfirmation"
                                                        v-bind:dataSource="selectedDatasource" v-bind:node="n"
                                                        ></entrypoint-content>
                                </v-tab-item>
                            </template>
                        </v-tabs>
                    </template>
                    <template v-else>
                        <div >
                            No node selected
                        </div>
                    </template>
                </div> -->
            </splitpanes>
        </div>

        <terminal
                :dataSource="selectedDatasource"
                v-if="selectedDatasource">
        </terminal>
    </div>
</template>

<script>
    import axios from 'axios';
    import EntrypointList from "./EntrypointList";
    import EntrypointContent from "./EntrypointContent";
    import Terminal from './Terminal.vue';
    import DataSource from "../models/DataSource";
    import Splitpanes from 'splitpanes'
    import EventBus from '../eventBus'
    import { mapState } from 'vuex'
    import _ from 'lodash';
    import { FETCH_DATASOURCE } from '../store/actions.type'

    export default {
        name: 'DataSourceList',
        components: {EntrypointList, EntrypointContent, Splitpanes, Terminal},

        computed: mapState({
            datasources: state => state.datasource.datasources
        }),

        data() {
            return {
                errors: [],
                selectedDatasourceId: null,
            }
        },

        methods: {
            showConfirmation: function (event) {
                this.$emit('display-modal', event);
            },

            refresh() {
                this.$store.dispatch(FETCH_DATASOURCE)
            },

            select: function (datasource) {
                this.selectedDatasourceId = datasource.id
            }
        },

        created() {
            this.refresh();
        }
    }
</script>

<style lang="scss" scoped>

    .splitpanes {
        // 74px = 64px (navbar) + 10px (margin-top)
        height: calc(100vh - 74px);
    }

    .v-tab {
        max-width: 100%;
    }

    .details-container {
        max-height: 100%;
        width: 100%;
    }

    .entrypoint-list-container {
        max-height: 100%;
        width: 100%;
    }

</style>
