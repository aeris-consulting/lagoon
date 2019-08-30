<template>
    <li v-bind:class="{ 'loading': loading}">
        <font-awesome-icon @click="toggleOpen()" class="icon-left" icon="angle-right"
                           v-if="node.hasChildren() && !open"/>
        <font-awesome-icon @click="toggleOpen()" class="icon-left" icon="angle-down" v-if="node.hasChildren() && open"/>
        <span @click="display()" class="name" v-bind:class="{ 'content': node.hasContent}">{{ node.name }}</span>
        <span class="childrenLength"
              v-if="node.hasChildren()">({{ node.length ? node.length : node.children.length }})
            <!-- <v-btn icon @click="add()" x-small>
              <font-awesome-icon icon="plus"/>
            </v-btn> -->
            <v-btn 
                icon @click="refresh()" x-small 
                v-if="node.hasChildren() && open">
              <font-awesome-icon icon="sync"/>
            </v-btn>
            <v-btn icon @click="copyChildrenList()" x-small
                v-if="node.hasChildren() && open">
              <font-awesome-icon icon="copy"/>
            </v-btn>
            <!-- <v-btn icon @click="deleteChildren()" x-small
                v-if="!dataSource.readonly">
              <font-awesome-icon icon="trash"/>
            </v-btn> -->
        </span>
        <span>
            <v-progress-circular
                v-if="loading"
                indeterminate
                :size="10"
                :width="2"
                color="primary"
            ></v-progress-circular>
        </span>
    </li>
</template>

<script>
    import Node from '../models/Node';
    import EntrypointChildren from './EntrypointChildren';
    import EventBus from '../eventBus'
    import Vue from 'vue';

    export default {
        name: 'entrypoint',

        props: {
            node: Object,
            dataSource: Object,
        },

        data() {
            return {
                open: false,
                loading: false,
                childrenComponent: null,
                contentLoaded: false,
                entrypointChildrenClass: Vue.extend(EntrypointChildren),
            }
        },

        methods: {
            showConfirmation: function (event) {
                this.$emit('display-modal', event);
            },

            showSnakebar: function (event) {
                EventBus.$emit('display-snakebar', event);
            },

            display: function () {
                if (this.node.hasContent) {
                    this.dataSource.selectNode(this.node);
                }
            },

            toggleOpen: function () {
                if (this.node.hasChildren()) {
                    if (!this.open) {
                        if (this.childrenComponent === null) {
                            this.refreshChildren();
                        } else if (this.childrenComponent !== null) {
                            this.open = !this.open;
                            this.childrenComponent.visible = true;
                        }
                    } else if (this.childrenComponent !== null) {
                        this.open = !this.open;
                        this.childrenComponent.visible = false;
                    }
                }
            },

            copyChildrenList: function () {
                let fullName = this.node.getFullName();
                let value;
                // eslint-disable-next-line
                console.log(this.node.children);

                this.node.children.forEach((v, k) => {
                    let nodeFullName = fullName + ':' + k;
                    if (value) {
                        value += "\r\n" + nodeFullName;
                    } else {
                        value = nodeFullName;
                    }
                });

                if (value) {
                    this.$copyText(value).then(function () {
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

            deleteChildren: function () {
                EventBus.$emit('display-snakebar', {
                    message: 'Not yet implemented'
                });
                /*this.$emit('display-modal', {
                    message: 'Are you sure you want to delete all the children?',
                    yesHandler: () => {
                        this.dataSource.deleteEntrypointChildren(this.node);
                    }, noHandler: () => {
                    }
                });*/
            },

            close: function () {
                this.open = false;
                this.node.clear();
                this.$el.removeChild(this.childrenComponent.$el);
                this.childrenComponent = null;
            },

            refreshChildren: function () {
                this.loading = true;
                let self = this;
                    
                if (this.node.children !== null) {
                    this.node.children.clear();
                }

                this.dataSource.listEntrypoints(this.node.getFullName(), this.node.level, this.node.level, receivedValues => {
                    receivedValues.forEach(value => {
                        self.node.addChildNode(new Node(value.path, value.length, value.hasContent))
                    });

                    if (!this.childrenComponent) {
                        this.childrenComponent = new this.entrypointChildrenClass({
                            propsData: {
                                children: this.node.children.values(),
                                dataSource: this.dataSource,
                            }
                        });
                        this.childrenComponent.$mount();
                        this.$el.appendChild(this.childrenComponent.$el);
                        this.childrenComponent.$on(['display-modal'], this.showConfirmation);
                    } else {
                        this.childrenComponent.children = this.node.children.values();
                    }

                    this.loading = false;
                    this.open = true;
                }, () => {
                    this.loading = false;
                    this.open = false;
                });
            }
        },

        created() {
            this.node.component = this;
        },

        beforeDestroy() {
            this.node.component = null;
        }
    }
</script>

<style lang="scss" scoped>
    li {
        font-family: "Courier New";
        font-size: 13px;
        list-style-type: none;
        padding-left: 10px;
    }

    .name {
        margin-left: 5px;
        margin-right: 5px;
    }

    .content {
        cursor: pointer;

        &:hover {
            text-decoration: underline;
        }
    }

    .childrenLength {
        margin-left: 0px;
        margin-right: 5px;
    }

    li.loading {
        color: gray;
    }

    .icon-left {
        color: lightgray;
        font-size: 10px;
        cursor: pointer;
    }

    .icon-right {
        margin-right: 4px;
        color: lightgray;
        font-size: 10px;
        cursor: pointer;
    }

</style>
