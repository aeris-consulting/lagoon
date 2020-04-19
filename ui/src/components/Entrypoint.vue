<template>
    <div id="tree-node">
        <div class="entrypoint"
            @mouseover="hover = true"
            @mouseleave="hover = false">
            <font-awesome-icon @click="toggle()" class="toggle-icon" icon="angle-right"
                v-if="node.hasChildren && !isOpen && !loading"/>
            <font-awesome-icon @click="toggle()" class="toggle-icon" icon="angle-down"
                v-if="node.hasChildren && isOpen && !loading"/>
            <v-progress-circular
                :size="10"
                :width="2"
                color="primary"
                indeterminate
                v-if="loading"
            ></v-progress-circular>
            <span @click="display()" :class="{ 'content': node.hasContent }">
                {{ node.path }}
                <span v-if="node.hasChildren">({{node.length}})</span>
            </span>
            <span class="entrypoint-actions" v-if="hover">
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
                        @click="deleteChildren(node)"
                        icon v-if="node.hasChildren && !readonly"
                        x-small>
                    <font-awesome-icon icon="trash"/>
                </v-btn>
            </span>
        </div>
        <div v-if="children && children.length > 0 && isOpen" class="entrypoint-children">
            <entrypoint v-for="child in children" :key="child.path" :node="child" :filter="filter">
            </entrypoint>
        </div>
    </div>
</template>

<script>
    import EventBus from '../eventBus';
    import {DELETE_NODE, FETCH_ENTRY_POINTS, SELECT_NODE} from '../store/actions.type';
    import {UNSELECT_NODE} from '../store/mutations.type';

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
                loading: false,
                hover: false
            }
        },

        methods: {
            toggle() {
                if (this.children === null && !this.isOpen) {
                    this.fetchEntryPoints()
                }
                this.isOpen = !this.isOpen
            },

            display() {
                if (this.node.hasContent) {
                    this.$store.dispatch(SELECT_NODE, this.node)
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

            copyChildrenList() {
                let valueToCopy;
                this.children.forEach((v) => {
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
                    filter: `*${this.filter}*`,
                    entrypointPrefix: this.node.fullPath,
                    minLevel: this.node.level + 1,
                    maxLevel: this.node.level + 1,
                }).then(data => {
                    this.children = Object.freeze([...data.map(n => {
                        n.hasChildren = n.length > 0
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
            this.$store.subscribe((mutation) => {
                if (mutation.type === UNSELECT_NODE) {
                    const deletedNode = mutation.payload
                    if (this.children && this.children.length) {
                        let deletedChildNode = this.children.find(c => c.fullPath === deletedNode.fullPath);
                        if (deletedChildNode) {
                            this.fetchEntryPoints();
                        }
                    }
                }
            })
        }
    }
</script>

<style lang="scss">
    #tree-node {    
        font-family: "Ubuntu Mono";
        .entrypoint-children {
            margin-left: 20px;
        }
        .toggle-icon {
            cursor: pointer;
            color: rgb(172, 172, 172);
            font-size: 13px;
        }
        .content {
            cursor: pointer;
            &:hover {
                text-decoration: underline;
            }
        }
    }
</style>
