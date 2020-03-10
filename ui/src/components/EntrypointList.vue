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
            v-if="nodes && nodes.length > 0">
            <v-treeview
                :items="nodes"
                :load-children="fetchEntryPoints"
                dense
                transition
            >
                <template v-slot:label="{ item }">
                    <span @click="display(item)" :class="{ 'content': item.hasContent }">{{item.path}}</span>
                </template>
            </v-treeview>
        </div>
    </div>
</template>

<script>
    import EntrypointChildren from "./EntrypointChildren";
    import Node from "../models/Node";
    import { FETCH_ENTRY_POINTS, SELECT_NODE, UNSELECT_NODE } from '../store/actions.type';

    export default {
        name: 'EntrypointList',
        components: {EntrypointChildren},

        props: {
            datasourceId: String,
        },

        computed: {
            datasource() {
                return this.$store.getters.getSelectedDatasource()
            }
        },

        data() {
            return {
                filter: '',
                loading: false,
                nodes: []
            }
        },

        methods: {
            showConfirmation: function (event) {
                this.$emit('display-modal', event);
            },

            display(node) {
                if (node.hasContent) {
                    this.$store.dispatch(SELECT_NODE, node)
                }
            },

            async fetchEntryPoints(node) {
                node.children = [];
                return this.$store.dispatch(FETCH_ENTRY_POINTS, {
                    filter: `${this.filter},${node.fullPath}*`,
                    entrypointPrefix: node.path,
                    minLevel: node.level + 1,
                    maxLevel: node.level + 1,
                }).then(data => {
                    node.children.push(...data.map(n => {
                        if (n.length > 0) {
                            n.children = []
                        }
                        n.name = n.path
                        n.fullPath = node.fullPath + ':' + n.path
                        n.level = node.level + 1
                        return n;
                    }))
                })
            },

            async refresh() {
                this.loading = true;
                let self = this;
                this.nodes = []
                const data = await this.$store.dispatch(FETCH_ENTRY_POINTS, {
                    filter: this.filter,
                    entrypointPrefix: null,
                    minLevel: 0,
                    maxLevel: 0,
                });

                this.loading = false;
                this.nodes = data.map(n => {
                    if (n.length > 0) {
                        n.children = []
                    }
                    n.name = n.path
                    n.fullPath = n.path
                    n.level = 0
                    return n;
                });
            },

            dismissErrorMessage: function(errorIndex) {
                this.dataSource.errors.splice(errorIndex, 1);
            }
        },

        created() {
            this.$store.subscribe((mutation, state) => {
                if (mutation.type === UNSELECT_NODE) {
                    const deletedNode = mutation.payload
                    if (deletedNode.level === 0) {
                        this.refresh();
                    } else {
                        // finding the parent node of the deleted node
                        let parentNode = null;
                        let treeToSearch = this.nodes;
                        deletedNode.fullPath.split(':').slice(0, -1).forEach((path, idx) => {
                            parentNode = treeToSearch.find(n => n.name === path);
                            if (parentNode && parentNode.children) {
                                treeToSearch = parentNode.children;
                            }
                        });

                        this.fetchEntryPoints(parentNode);
                    }
                }
            })
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

    .content {
        cursor: pointer;

        &:hover {
            text-decoration: underline;
        }
    }
</style>
