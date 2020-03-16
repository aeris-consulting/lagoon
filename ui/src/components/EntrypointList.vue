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
            <v-treeview
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
            </v-treeview>
        </div>
    </div>
</template>

<script>
    import EventBus from '../eventBus';
    import { FETCH_ENTRY_POINTS, SELECT_NODE, DELETE_NODE } from '../store/actions.type';
    import { UNSELECT_NODE } from '../store/mutations.type';

    export default {
        name: 'EntrypointList',

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

            deleteChildren(node) {
                EventBus.$emit('display-modal', {
                    message: 'Are you sure you want to delete the content?',
                    yesHandler: () => {
                        this.$store.dispatch(DELETE_NODE, node)
                    }, noHandler: () => {}
                });
            },

            copyChildrenList(node) {
                let valueToCopy;
                node.children.forEach((v) => {
                    if (valueToCopy) {
                        valueToCopy += "\r\n" + v.fullPath;
                    } else {
                        valueToCopy = v.fullPath;
                    }
                });

                if (valueToCopy) {
                    this.$copyText(valueToCopy).then(function () {
                        EventBus.$emit('display-snakebar', {
                            message: 'The list of direct children was copied to your clipboard'
                        });
                    }, function () {
                        EventBus.$emit('display-snakebar', {
                            message: 'The list of direct children could not be copied to your clipboard!!!'
                        });
                    })
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
                        if (n.length > 0) {
                            n.children = []
                        }
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
