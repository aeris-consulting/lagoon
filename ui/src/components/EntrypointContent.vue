<template>
    <v-container fluid>
        <v-row class="button-bar">
            <v-col cols="8">
                <v-btn @click="refresh()" icon large>
                    <font-awesome-icon icon="sync"/>
                </v-btn>
                <v-btn @click="observe()" icon large v-if="!observing">
                    <font-awesome-icon icon="play"/>
                </v-btn>
                <v-btn @click="stopObserve()" icon large v-if="observing">
                    <font-awesome-icon icon="stop"/>
                </v-btn>
                <input type="number" class="frequency-input" v-model="observationFrequency"/> seconds
                <v-chip
                        v-if="lastRefresh"
                        class="mr-2">
                    <v-tooltip bottom>
                        <template v-slot:activator="{ on }">
                            <v-icon left v-on="on">mdi-clock</v-icon>
                        </template>
                        <span>Last refresh time</span>
                    </v-tooltip>
                    <template>
                        {{ lastRefresh.toISOString() }}
                    </template>
                </v-chip>
            </v-col>
            <v-col cols="4">
                <v-row justify="end">
                    <v-btn @click="edit()" icon large v-if="!datasource.readonly">
                        <font-awesome-icon icon="edit"/>
                    </v-btn>
                    <v-btn @click="deleteNode()" icon large v-if="!datasource.readonly">
                        <font-awesome-icon icon="trash"/>
                    </v-btn>
                </v-row>
            </v-col>
        </v-row>

        <template v-if="isLoadingContent">
            <v-row>
                <v-progress-circular
                        class="ml-2"
                        color="primary"
                        indeterminate
                ></v-progress-circular>
            </v-row>
        </template>
        <template v-else>
            <div class="entrypoint-content-panel">
                <v-row>
                    <v-col cols="6">
                        <v-chip
                                @click="copyKey"
                                class="mr-2">
                            <v-tooltip bottom>
                                <template v-slot:activator="{ on }">
                                    <v-icon left v-on="on">mdi-key</v-icon>
                                </template>
                                <span>Name of node (click to copy)</span>
                            </v-tooltip>
                            {{ node.fullPath }}
                        </v-chip>
                    </v-col>
                    <v-col cols="6" style="text-align: right">
                        <template v-if="nodeDetails.info">
                            <v-chip
                                    class="mr-2">
                                <v-tooltip bottom>
                                    <template v-slot:activator="{ on }">
                                        <v-icon left v-on="on">mdi-shape</v-icon>
                                    </template>
                                    <span>Type</span>
                                </v-tooltip>
                                {{ nodeDetails.info.type.toLowerCase() }}
                            </v-chip>
                            <v-chip
                                    class="mr-2">
                                <v-tooltip bottom>
                                    <template v-slot:activator="{ on }">
                                        <v-icon left v-on="on">mdi-ruler</v-icon>
                                    </template>
                                    <span>Length</span>
                                </v-tooltip>
                                {{ nodeDetails.info.length }}
                            </v-chip>
                            <template v-if="timeToLive">
                                <v-chip
                                        class="mr-2">
                                    <v-tooltip bottom>
                                        <template v-slot:activator="{ on }">
                                            <v-icon left v-on="on">mdi-timer-sand</v-icon>
                                        </template>
                                        <span>Time to live</span>
                                    </v-tooltip>
                                    {{ timeToLive }}
                                </v-chip>
                            </template>
                        </template>
                    </v-col>
                </v-row>

                <div class="content mt-2" v-if="nodeDetails.content && nodeDetails.info">
                    <h4>Content</h4>
                    <json-viewer
                            :expand-depth=this.defaultExpandDepth
                            :value="nodeDetails.content.data | parseIfIsJson"
                            copyable>
                    </json-viewer>
                </div>
            </div>

        </template>
    </v-container>
</template>
<script>
    import EventBus from '../eventBus';
    import JsonHelper from '../helpers/jsonHelper';
    import NodeHelper from '../helpers/nodeHelper';
    import {DELETE_NODE, FETCH_NODE_DETAILS} from '../store/actions.type';
    import {ADD_ERROR} from "../store/mutations.type";

    const humanizeDuration = require('humanize-duration');

    export default {
        name: 'EntrypointContent',

        props: {
            node: Object,
        },

        data() {
            return {
                nodeDetails: null,
                defaultExpandDepth: 1,
                observing: false,
                observationFrequency: 10,
                lastRefresh: null,
                isLoadingContent: true,
                format: null
            }
        },

        methods: {
            refresh: async function () {
                this.lastRefresh = new Date();
                this.isLoadingContent = true;

                this.$store.dispatch(FETCH_NODE_DETAILS, this.node)
                    .then(nodeDetails => {
                        this.nodeDetails = nodeDetails;
                        if (this.nodeDetails.info.type == 'SCORED_SET') {
                            const data = this.nodeDetails.content.data;
                            this.nodeDetails.content.data = new Map();
                            for (let value of data) {
                                this.nodeDetails.content.data[value.score] = value.values
                            }
                            this.defaultExpandDepth = 2
                        } else if (this.nodeDetails.info.type == 'HASH') {
                            const data = this.nodeDetails.content.data;
                            this.nodeDetails.content.data = new Map();
                            for (let value of data) {
                                this.nodeDetails.content.data[value.key] = value.value
                            }
                            this.defaultExpandDepth = 2
                        } else if (this.nodeDetails.info.type == 'VALUE') {
                            // In case the value is a JSON, we open wide enough to be able to read it.
                            this.defaultExpandDepth = 5
                        }
                        this.isLoadingContent = false;
                    }).catch((e) => {
                        this.loading = false;
                        this.$store.commit(ADD_ERROR, e);
                    });
            },

            observe: function () {
                this.observing = true;
                this.refresh();
                this.scheduleNextRefresh();
            },

            stopObserve: function () {
                this.observing = false;
                if (this.observationFlag) {
                    clearTimeout(this.observationFlag);
                }
            },

            scheduleNextRefresh: function () {
                this.observationFlag = setTimeout(() => {
                    if (this.observing) {
                        this.refresh();
                        this.scheduleNextRefresh();
                    }
                }, this.observationFrequency * 1000);
            },

            edit: function () {
                EventBus.$emit('display-snakebar', {
                    message: 'Not yet implemented'
                });
            },

            deleteNode: function () {
                EventBus.$emit('display-modal', {
                    message: 'Are you sure you want to delete the content?',
                    yesHandler: () => {
                        this.stopObserve();

                        this.$store.dispatch(DELETE_NODE, this.node).then(() => {
                            // Update the tree.
                            this.node.hasContent = false;
                            let nodeToRefresh = this.node.parent;
                            while (nodeToRefresh != null) {
                                nodeToRefresh.length -= 1;
                                nodeToRefresh.hasChildren = nodeToRefresh.length > 0;
                                NodeHelper.removeFromParentIfEmpty(nodeToRefresh);
                                nodeToRefresh = nodeToRefresh.parent;
                            }
                        }).catch((e) => {
                            this.$store.commit(ADD_ERROR, e);
                        })
                    }, noHandler: () => {
                    }
                });
            },

            copyKey: function () {
                this.$copyText(this.node.fullPath).then(function () {
                    EventBus.$emit('display-snakebar', {
                        message: 'The key was copied to your clipboard'
                    });
                }, function () {
                    EventBus.$emit('display-snakebar', {
                        message: 'The key could not be copied to your clipboard'
                    });
                })
            },
        },

        computed: {
            datasource() {
                return this.$store.getters.getSelected
            },
            timeToLive: function () {
                if (this.nodeDetails && this.nodeDetails.info.timeToLive && this.nodeDetails.info.timeToLive > 0) {
                    return humanizeDuration(this.nodeDetails.info.timeToLive, {units: ['d', 'h', 'm', 's']})
                }
                return null
            }
        },

        filters: {
            parseIfIsJson: function (value) {
                if (value && value.length && value.length === 1) {
                    if (JsonHelper.isJson(value[0])) {
                        return JSON.parse(value);
                    }
                }
                return value
            }
        },

        created() {
            this.refresh();
        },

        beforeDestroy() {
            this.stopObserve();
        }
    }
</script>

<style lang="scss" scoped>
    input.frequency-input {
        width: 40px !important;
        margin-left: 10px;
    }


    h2 {
        font-size: 1.2em;
    }

    h3 {
        font-size: 1.1em;
    }

    .button-bar {
        font-family: "Ubuntu Mono";
        background-color: #f5f5f5;
        border-radius: .25rem;
        margin-bottom: 10px;
        font-size: 14px;

        .icon {
            margin: 0;
            margin-right: 5px;
            cursor: pointer;
        }

        .observation-frequency {
            width: 30px;
            font-size: 12px;
            text-align: right;
        }

        .space {
            min-width: 20px;
        }
    }

    .entrypoint-content-panel {
        font-family: "Ubuntu Mono";
    }

    .info {
        margin-top: 20px;
        margin-bottom: 30px;
    }

    .content-data td {
        font-family: "Courier New";
        font-size: 14px;
        text-align: left;
    }

</style>