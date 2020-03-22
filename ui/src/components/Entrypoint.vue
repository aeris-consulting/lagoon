<template>
    <div id="tree-node">
        <div class="entrypoint">
            <font-awesome-icon @click="toggle()" class="icon-left" icon="angle-right"
                v-if="node.hasChildren && !isOpen && !loading"/>
            <font-awesome-icon @click="toggle()" class="icon-left" icon="angle-down"
                v-if="node.hasChildren && isOpen && !loading"/>
            <v-progress-circular
                :size="10"
                :width="2"
                color="primary"
                indeterminate
                v-if="loading"
            ></v-progress-circular>
            {{ node.path }} 
            <span v-if="node.hasChildren">({{node.length}})</span>
            <span class="entrypoint-actions">
                <v-btn
                    icon
                    @click="fetchEntryPoints(node)" v-if="node.hasChildren && isOpen"
                    x-small>
                    <font-awesome-icon icon="sync"/>
                </v-btn>
                <v-btn 
                    icon
                    @click="copyChildrenList(node)" v-if="node.hasChildren && isOpen"
                    x-small>
                    <font-awesome-icon icon="copy"/>
                </v-btn>
                <v-btn 
                    icon
                    @click="deleteChildren(node)" v-if="!readonly"
                    x-small>
                    <font-awesome-icon icon="trash"/>
                </v-btn>
            </span>
        </div>
        <div v-if="children && children.length > 0 && isOpen" class="entrypoint-children">
            <entrypoint v-for="(child, index) in children" :key="index" :node="child" :filter="filter">
            </entrypoint>
        </div>
        <!-- <div>
            LOAD MORE
        </div> -->
    </div>
</template>

<script>
    import EventBus from '../eventBus';
    import { FETCH_ENTRY_POINTS, SELECT_NODE, DELETE_NODE } from '../store/actions.type';
    // import { UNSELECT_NODE } from '../store/mutations.type';

    export default {
        name: 'entrypoint',

        props: {
            node: Object,
            filter: String,
            readonly: Boolean
        },

        data() {
            return {
                isOpen: false,
                children: null,
                loading: false
            }
        },

        methods: {
            toggle() {
                if (this.children === null && !this.isOpen) {
                    this.fetchEntryPoints()
                }
                this.isOpen = !this.isOpen
            },

            deleteChildren(node) {
                EventBus.$emit('display-modal', {
                    message: 'Are you sure you want to delete the content?',
                    yesHandler: () => {
                        this.$store.dispatch(DELETE_NODE, node)
                    }, noHandler: () => {}
                });
            },

            copyChildrenList() {
                let valueToCopy;
                this.node.children.forEach((v) => {
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

            fetchEntryPoints() {
                this.loading = true
                this.$store.dispatch(FETCH_ENTRY_POINTS, {
                    filter: `${this.filter},${this.node.fullPath}:*`,
                    entrypointPrefix: this.node.path,
                    minLevel: this.node.level + 1,
                    maxLevel: this.node.level + 1,
                }).then(data => {
                    this.children = Object.freeze([...data.map(n => {
                        n.hasChildren = n.length > 0 ? true : false
                        n.name = n.path
                        n.fullPath = this.node.fullPath + ':' + n.path
                        n.level = this.node.level + 1
                        return n;
                    })]);
                    this.loading = false
                })
            },
        },

        created() {
        }
    }
</script>

<style lang="scss">
    .entrypoint-children {
        margin-left: 20px;
    }

    .entrypoint {
        .entrypoint-actions {
            display: none;
        }
        &:hover {
            .entrypoint-actions {
                display: inline-block;
            }            
        }
    }
</style>
