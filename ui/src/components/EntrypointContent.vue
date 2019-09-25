<template>
    <v-container fluid>
        <v-row>
            <v-col cols="8">
                <v-btn icon @click="refresh()" small>
                    <font-awesome-icon icon="sync"/>
                </v-btn>
                <v-btn icon @click="observe()" small v-if="!observing">
                    <font-awesome-icon icon="play"/>
                </v-btn>
                <v-btn icon @click="stopObserve()" small v-if="observing">
                    <font-awesome-icon icon="stop"/>
                </v-btn>
                <input type="number" class="frequency-input" v-model="observationFrequency"/> seconds
            </v-col>
            <v-col cols="4">
                <v-row justify="end">
                    <v-btn icon @click="edit()" small v-if="!dataSource.readonly">
                        <font-awesome-icon icon="edit"/>
                    </v-btn>
                    <v-btn icon @click="erase()" small v-if="!dataSource.readonly">
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
            <v-row>
                <v-col cols="12">
                    <v-chip
                            @click="copyKey"
                            class="mr-2">
                        <v-tooltip bottom>
                            <template v-slot:activator="{ on }">
                                <v-icon left v-on="on">mdi-key</v-icon>
                            </template>
                            <span>Type of node</span>
                        </v-tooltip>
                        {{ node.getFullName() }}
                    </v-chip>
                </v-col>
            </v-row>
            <v-row>
                <v-col cols="12">
                    <v-chip
                            class="mr-2">
                        <v-tooltip bottom>
                            <template v-slot:activator="{ on }">
                                <v-icon left v-on="on">mdi-clock</v-icon>
                            </template>
                            <span>Last refresh time</span>
                        </v-tooltip>
                        <template v-if="lastRefresh">
                            {{ lastRefresh.toISOString() }}
                        </template>
                    </v-chip>
                    <template v-if="node.info">
                        <v-chip
                                class="mr-2">
                            <v-tooltip bottom>
                                <template v-slot:activator="{ on }">
                                    <v-icon left v-on="on">mdi-shape</v-icon>
                                </template>
                                <span>Type of node</span>
                            </v-tooltip>
                            {{ node.info.type.toLowerCase() }}
                        </v-chip>
                        <v-chip
                                class="mr-2">
                            <v-tooltip bottom>
                                <template v-slot:activator="{ on }">
                                    <v-icon left v-on="on">mdi-ruler</v-icon>
                                </template>
                                <span>length of value</span>
                            </v-tooltip>
                            {{ node.info.length }}
                        </v-chip>
                    </template>
                </v-col>
            </v-row>

            <div class="content mt-2" v-if="node.content && node.info">
                <h4>Content</h4>
                <div v-if="node.info.type == 'HASH'">
                    <v-simple-table dense>
                        <thead>
                        <tr>
                            <td>Field</td>
                            <td>Value</td>
                        </tr>
                        </thead>
                        <tbody>
                        <tr :key="k" class="content-data" v-for="(v,k) in node.content.data[0]">
                            <td>{{ k }}</td>
                            <td>{{ v }}</td>
                        </tr>
                        </tbody>
                    </v-simple-table>
                </div>

                <div class="content-data" v-else>
                    <json-viewer
                            :expand-depth=3
                            :value="node.content.data | parseIfIsJson"
                            boxed
                            copyable
                            sort>
                    </json-viewer>
                </div>
            </div>

        </template>
    </v-container>
</template>
<script>
    import EventBus from '../eventBus';
    import JsonHelper from '../helpers/jsonHelper';

    export default {
        name: 'EntrypointContent',

        props: {
            dataSource: Object,
            node: Object,
        },

        data() {
            return {
                observing: false,
                observationFrequency: 10,
                lastRefresh: null,
                isLoadingContent: false,
            }
        },

        methods: {
            refresh: function () {
                let self = this;
                this.isLoadingContent = true;
                this.dataSource.refreshNodeDetails(this.node).then(() => {
                    this.isLoadingContent = false;
                    self.lastRefresh = new Date();
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

            erase: function () {
                this.$emit('display-modal', {
                    message: 'Are you sure you want to delete the content?',
                    yesHandler: () => {
                        this.dataSource.deleteEntrypoint(this.node);
                    }, noHandler: () => {
                    }
                });
            },

            copyKey: function () {
                this.$copyText(this.node.getFullName()).then(function () {
                    EventBus.$emit('display-snakebar', {
                        message: 'The key was copied to your clipboard'
                    });
                }, function () {
                    EventBus.$emit('display-snakebar', {
                        message: 'The key could not be copied to your clipboard'
                    });
                })
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
            let node = this.node;
            node.contentComponent = self;
            this.refresh();
        },

        beforeDestroy() {
            this.stopObserve();
            this.node.contentComponent = null;
        }
    }
</script>

<style lang="scss" scoped>
    input.frequency-input {
        width: 40px !important;
        margin-left: 10px;
    }

    .entrypoint-content {

        width: 100%;

        h2 {
            font-size: 1.2em;
        }

        h3 {
            font-size: 1.1em;
        }

        .button-bar {
            background-color: lightgrey;
            padding: 5px;
            padding-left: 15px;
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

        .info {
            margin-top: 20px;
            margin-bottom: 30px;
        }

        .content-timestamp-data, .info-data, .content-data {
            font-family: "Courier New";
            font-size: 14px;
            text-align: left;
        }

    }
</style>
