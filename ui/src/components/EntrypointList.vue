<template>
    <div id="entrypoints">
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
            

            <entrypoint v-for="(node, index) in nodes" :key="index" :node="node" :filter="filter" :readonly="datasource.readonly">
            </entrypoint>

            <!-- <button @click="fetchEntryPoints(nodes[0])">
                CLICK ME
            </button>
            <div v-for="n in nodes">
                {{n.fullPath}}
                <div v-if="n.children && n.children.length > 0">
                    <div v-for="c in n.children">
                        {{c.fullPath}}
                    </div>
                </div>
            </div> -->

            <!-- <v-treeview
                :items="nodes"
                :load-children="fetchEntryPoints"
                dense
                transition
            >
                <template v-slot:label="{ item: node, open }">
                    <span @click="display(node)" :class="{ 'content': node.hasContent }">{{node.path}}</span>
                    <v-btn
                        icon
                        @click="fetchEntryPoints(node)" v-if="node.hasContent && open"
                        x-small>
                      <font-awesome-icon icon="sync"/>
                    </v-btn>
                    <v-btn 
                        icon
                        @click="copyChildrenList(node)" v-if="node.hasContent && open"
                        x-small>
                      <font-awesome-icon icon="copy"/>
                    </v-btn>
                    <v-btn 
                        icon
                        @click="deleteChildren(node)" v-if="!datasource.readonly"
                        x-small>
                      <font-awesome-icon icon="trash"/>
                    </v-btn>
                </template>
            </v-treeview> -->
        </div>
    </div>
</template>

<script>
    import EventBus from '../eventBus';
    import { FETCH_ENTRY_POINTS, SELECT_NODE, DELETE_NODE } from '../store/actions.type';
    import { UNSELECT_NODE } from '../store/mutations.type';
    import Entrypoint from './Entrypoint.vue';

    export default {
        name: 'EntrypointList',

        components: {
            Entrypoint
        },

        props: {
            datasourceId: String,
        },

        computed: {
            datasource() {
                return this.$store.getters.getSelected()
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
            display(node) {
                if (node.hasContent) {
                    this.$store.dispatch(SELECT_NODE, node)
                }
            },

            refresh() {
                this.loading = true;
                this.nodes = []
                this.$store.dispatch(FETCH_ENTRY_POINTS, {
                    filter: this.filter,
                    entrypointPrefix: null,
                    minLevel: 0,
                    maxLevel: 0,
                }).then(data => {
                    this.loading = false;
                    this.nodes = data.map(n => {
                        n.hasChildren = n.length > 0 ? true : false
                        n.name = n.path
                        n.fullPath = n.path
                        n.level = 0
                        return n;
                    });
                }).catch(() => {
                    this.loading = false;
                })
            },
        },

        created() {
            this.$store.subscribe((mutation) => {
                if (mutation.type === UNSELECT_NODE) {
                    const deletedNode = mutation.payload
                    if (deletedNode.level === 0) {
                        this.refresh();
                    } else {
                        // finding the parent node of the deleted node
                        let parentNode = null;
                        let treeToSearch = this.nodes;
                        deletedNode.fullPath.split(':').slice(0, -1).forEach((path) => {
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
