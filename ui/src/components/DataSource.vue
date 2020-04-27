<template>
    <div>
        <div v-if="selectedDatasourceId === null">
            <v-btn @click="refresh()" color="primary">Refresh data sources</v-btn>

            <div class="datasources" v-if="datasources && datasources.length > 0">
                <v-list>
                    <v-list-item
                            :key="datasource.id"
                            @click="select(datasource)"
                            v-for="datasource in datasources">
                        <v-list-item-content>
                            <v-list-item-title>{{ datasource.name }}
                                <template v-if="datasource.readonly">&nbsp;(Read-only)</template>
                            </v-list-item-title>
                            <v-list-item-subtitle v-if="datasource.description">{{ datasource.description }}
                            </v-list-item-subtitle>
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
                    <entrypoint-list></entrypoint-list>
                </div>
                <div class="details-container"
                     splitpanes-size="70">
                    <template v-if="selectedNodes.length > 0">
                        <v-tabs
                            v-model="activeNodeIndex"
                            show-arrows
                            background-color="primary"
                            dark
                            splitpanes-size="70">
                            <v-tab v-for="n in selectedNodes" :key="n.fullPath">
                                <span class="tab-title" :title="n.fullPath">
                                    {{ n.fullPath }}
                                </span>
                            </v-tab>
                        </v-tabs>
                        <v-tabs-items v-model="activeNodeIndex">
                            <v-tab-item v-for="n in selectedNodes" :key="n.fullPath">
                                <entrypoint-content
                                        :node="n"></entrypoint-content>
                            </v-tab-item>
                        </v-tabs-items>
                    </template>
                    <template v-else>
                        <div>
                            No node selected
                        </div>
                    </template>
                </div>
            </splitpanes>
        </div>

        <terminal
                v-if="selectedDatasourceId">
        </terminal>
    </div>
</template>

<script>
    import EntrypointList from "./EntrypointList";
    import EntrypointContent from "./EntrypointContent";
    import Terminal from './Terminal.vue';
    import Splitpanes from 'splitpanes'
    import {mapState} from 'vuex'
    import {FETCH_DATASOURCE, SELECT_DATASOURCE} from '../store/actions.type'
    import {ADD_SELECTED_NODE} from '../store/mutations.type'
    import EventBus from "../eventBus";

    export default {
        name: 'DataSourceList',
        components: {EntrypointList, EntrypointContent, Splitpanes, Terminal},

        computed: mapState({
            datasources: state => state.datasource.datasources,
            selectedDatasourceId: state => state.datasource.selectedDatasourceId,
            selectedNodes: state => state.datasource.selectedNodes
        }),

        data() {
            return {
                activeNodeIndex: null,
                errors: []
            }
        },

        methods: {
            refresh() {
                this.$store.dispatch(FETCH_DATASOURCE)
            },

            select: function (datasource) {
                this.$store.dispatch(SELECT_DATASOURCE, datasource.id)
                EventBus.$emit('datasource-set', {datasource: datasource})
            }
        },

        created() {
            this.refresh();

            this.$store.subscribe((mutation) => {
                if (mutation.type === ADD_SELECTED_NODE) {
                    const selectedNode = mutation.payload
                    this.activeNodeIndex = this.selectedNodes.findIndex(n => n.fullPath === selectedNode.fullPath)
                }
            })
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
