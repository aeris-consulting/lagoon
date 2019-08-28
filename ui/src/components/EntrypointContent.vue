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

        <v-list>
            <v-list-item v-if="lastRefresh !== null">
                <v-list-item-icon>
                    <v-icon>mdi-clock</v-icon>
                    <v-list-item-content>
                        <v-list-item-title>
                            {{ lastRefresh.toISOString() }}
                        </v-list-item-title>
                    </v-list-item-content>
                </v-list-item-icon>
            </v-list-item>
            <template v-if="node.info">
                <v-subheader>INFORMATION</v-subheader>
                <v-list-item>
                    <v-list-item-content>
                        <v-list-item-title>{{ node.info.type.toLowerCase() }}</v-list-item-title>
                        <v-list-item-subtitle>type of node</v-list-item-subtitle>
                    </v-list-item-content>
                </v-list-item>
                <v-list-item>
                    <v-list-item-content>
                        <v-list-item-title>{{ node.info.length }}</v-list-item-title>
                        <v-list-item-subtitle>length of value</v-list-item-subtitle>
                    </v-list-item-content>
                </v-list-item>
            </template>
        </v-list>

        <div class="content" v-if="node.content && node.info">
            <h3>Content</h3>
            <div v-if="node.info.type == 'HASH'">
                <table>
                    <thead>
                    <tr class="content-header">
                        <td>Field</td>
                        <td>Value</td>
                    </tr>
                    </thead>
                    <tbody>
                    <tr class="content-data" v-for="(v,k) in node.content.data[0]" :key="k">
                        <td>{{ k }}</td>
                        <td>{{ v }}</td>
                    </tr>
                    </tbody>
                </table>
            </div>

            <div class="content-data" v-else>
                <table>
                    <tbody>
                    <tr :key="i" v-for="(v, i) in node.content.data">
                        <td>{{ v }}</td>
                    </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </v-container>
</template>
<script>
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
            }
        },

        methods: {
            refresh: function () {
                this.dataSource.refreshNodeDetails(this.node);
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
                if (this.observing) {
                    this.refresh();
                    this.scheduleNextRefresh();
                }
            },

            edit: function () {
                alert('Not yet implemented');
            },

            erase: function () {
                this.$emit('display-modal', {
                    message: 'Are you sure you want to delete the content?',
                    yesHandler: () => {
                        this.dataSource.deleteEntrypoint(this.node);
                    }, noHandler: () => {
                    }
                });
            }
        },

        created() {
            this.dataSource.refreshNodeDetails(this.node);
            this.node.contentComponent = this;
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
