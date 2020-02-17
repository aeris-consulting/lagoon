<template>
    <modal
            :clickToClose="false"
            :height="300"
            :minHeight="100"
            :minWidth="100"
            :width="600"
            @opened="initTerminal"
            draggable=".dialog-header"
            name="terminal"
            resizable>
        <div class="dialog-header">
            Lagoon Terminal
            <v-menu offset-y v-if="selectedNode">
                <template v-slot:activator="{ on }">
                    <span class="cluster-node"
                          v-on="on">
                        - 
                        <span v-if="selectedNode.id === ''">Default</span>
                        <span v-else>{{ selectedNode.ip }} ({{selectedNode.role}})</span>
                    </span>
                </template>
                <v-list dense>
                    <v-list-item
                            @click="useDefaultNode()">
                        <v-list-item-title>Default</v-list-item-title>
                    </v-list-item>
                    <v-list-item
                            :key="index"
                            @click="changeNode(item)"
                            v-for="(item, index) in clusterNodesInfo"
                    >
                        <v-list-item-title>{{ item.name }} ({{item.role}})</v-list-item-title>
                    </v-list-item>
                </v-list>
            </v-menu>
            <button
                    @click="$modal.hide('terminal')"
                    class="close-button">
                <font-awesome-icon :icon="['fa', 'times']"/>
            </button>
        </div>
        <div
                class="terminal">
        </div>
    </modal>
</template>

<script>
    var $ = require('jquery');
    require('jquery.terminal');
    import EventBus from '../eventBus';

    const defaultNode = {
        id: ''
    }

    export default {
        name: 'terminal',
        components: {},
        props: {
            dataSource: Object
        },
        data() {
            return {
                clusterNodesInfo: [],
                selectedNode: null
            }
        },
        methods: {
            useDefaultNode() {
                this.selectedNode = defaultNode
            },
            changeNode(clusterNode) {
                this.selectedNode = clusterNode
            },
            initTerminal() {
                const self = this;
                $('.terminal').terminal(function (command) {
                    if (command !== '') {
                        const commands = command.split(' ');
                        const nodeId = (self.selectedNode && self.selectedNode.id) ? self.selectedNode.id : ''
                        self.dataSource.executeCommand(commands, nodeId).then(response => {
                            echoDataToTerminal(this, response.data);
                        }).catch(e => {
                            this.echo(String(e));
                        });
                    }
                }, {
                    greetings: 'Lagoon redis teminal',
                    prompt: 'redis> '
                });
            },
        },
        mounted() {
            EventBus.$on('open-terminal', () => {
                this.$modal.show('terminal');
            });
            this.dataSource.getClusterNodes().then((clusterNodesInfo) => {
                this.clusterNodesInfo = clusterNodesInfo;
                if (clusterNodesInfo.length > 0) {
                    this.selectedNode = defaultNode
                }
            })
        }
    }

    function echoDataToTerminal(terminal, data) {
        if (Array.isArray(data)) {
            data.forEach(row => {
                echoDataToTerminal(terminal, row);
            })
        } else {
            terminal.echo(String(data));
        }
    }
</script>

<style lang="scss" scoped>
    @import '../../node_modules/jquery.terminal/css/jquery.terminal.min.css';
    @import '../assets/_variables.scss';

    .terminal {
        height: calc(100% - 28px);
    }

    .dialog-header {
        background-color: $blue;
        color: #fff;
        padding: 2px 10px 2px 10px;
        cursor: move;
    }

    .close-button {
        float: right;
    }

    .v--modal-overlay[data-modal="terminal"] {
        pointer-events: none;
        background: transparent;
    }

    .cluster-node {
        cursor: pointer;
        font-style: italic;
        font-size: 0.8em;
    }
</style>