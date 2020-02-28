<template>
    <div id="entrypoints">
        <div class="alerts-container">
            <!-- <v-alert
                :key="i" class="errors" v-for="(error, i) in dataSource.errors"
                :value="true"
                @input="dismissErrorMessage(i)"
                border="left"
                close-text="Close Alert"
                type="error"
                dark
                dismissible>
                {{ error.message }}
            </v-alert> -->
        </div>
        <div class="filter-panel">
            <div>
                <div class="filter-container">
                    <v-text-field
                        v-model="filter"
                        label="Filter"
                    ></v-text-field>
                </div>
                <v-btn class="" color="primary" @click="refresh()">List</v-btn>
                <v-progress-circular
                    class="loading-circle"
                    v-if="loading"
                    indeterminate
                    color="green">
                </v-progress-circular>
            </div>
        </div>

        <div class="entrypoint-children-panel" 
            v-if="firstLevelNodes && firstLevelNodes.length > 0">
            <entrypoint-children @display-modal="showConfirmation"
                                 v-bind:children="root.children.values()"
                                 v-bind:dataSource="dataSource"></entrypoint-children>
        </div>
    </div>
</template>

<script>
    import EntrypointChildren from "./EntrypointChildren";
    import Node from "../models/Node";
    import { FETCH_ENTRY_POINTS } from '../store/actions.type';

    export default {
        name: 'EntrypointList',
        components: {EntrypointChildren},

        props: {
            datasourceId: String,
        },

        computed: {
            datasource() {
                return this.$store.getters.getDataSourceById(this.datasourceId)
            }
        },

        data() {
            return {
                filter: '',
                loading: false,
                firstLevelNodes: []
            }
        },

        methods: {
            showConfirmation: function (event) {
                this.$emit('display-modal', event);
            },

            async refresh() {
                this.loading = true;
                let self = this;
                const data = await this.$store.dispatch(FETCH_ENTRY_POINTS, {
                    id: this.datasourceId,
                    filter: this.filter,
                    entrypointPrefix: null,
                    minLevel: 0,
                    maxLevel: 0,
                });

                this.firstLevelNodes = data;
                
                // this.dataSource.listEntrypoints(null, 0, 0, receivedValues => {
                //     receivedValues.forEach(value => {
                //         self.root.addChildNode(new Node(value.path, value.length, value.hasContent))
                //     });
                //     self.dataSource.status = null;
                // }, () => {
                //     self.dataSource.status = null;
                // }, () => {
                //     self.dataSource.status = null;
                // });
            },

            dismissErrorMessage: function(errorIndex) {
                this.dataSource.errors.splice(errorIndex, 1);
            }
        }
    }
</script>

<style lang="scss" scoped>
    .filter-container {
        margin-right: 15px;
        width: 200px;
        display: inline-block;
    }

    .alerts-container {
        position: fixed;
        z-index: 9999;
        top: 80px;
        right: 20px;
    }

    div#entrypoints {
        text-align: left;
        margin-right: 10px;
        position: relative;
        margin: 0 auto;
        height: 100%;

        .filter-panel {
            height: 70px;
            width: 100%;
        }

        .entrypoint-children-panel {
            overflow-x: auto;
            top: 70px;
            left: 0;
            right: 0;
            bottom: 0;
            position: absolute;
        }
    }

    div#entrypoints div.data {
        text-align: left;
    }

    .loading-circle {
        margin-left: 10px;
    }
</style>
