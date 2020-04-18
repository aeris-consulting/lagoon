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
                        n.hasChildren = n.length > 0
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
