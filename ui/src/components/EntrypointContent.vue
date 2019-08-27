<template>
    <div class="entrypoint-content">
        <div v-if="node">
            <h2>{{ node.getFullName() }}</h2>
            <div>
                <div class="button-bar row justify-content-between">
                    <div class="col-6">
                        <font-awesome-icon @click="refresh()" class="icon" icon="sync"/>
                        <font-awesome-icon @click="observe()" class="icon" icon="play" v-if="!observing"/>
                        <font-awesome-icon @click="stopObserve()" class="icon" icon="stop" v-if="observing"/>
                        <input class="observation-frequency" v-model="observationFrequency"> seconds
                    </div>
                    <div class="col-4">
                        <font-awesome-icon @click="edit()" class="icon" icon="edit" v-if="!dataSource.readonly"/>
                        <font-awesome-icon @click="erase()" class="icon" icon="trash" v-if="!dataSource.readonly"/>
                    </div>
                </div>
            </div>

            <div class="content-timestamp" v-if="lastRefresh !== null">
                <h3>Data timestamp</h3>
                <div class="content-timestamp-data">
                    {{ lastRefresh.toISOString() }}
                </div>
            </div>

            <div class="info" v-if="node.info">
                <h3>Information</h3>
                <div class="info-data">
                    <span>Type: {{ node.info.type.toLowerCase() }}</span>
                    <br/>
                    <span>Length: {{ node.info.length }}</span>
                </div>
            </div>

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
        </div>
        <div v-else>No node selected</div>
    </div>
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
                this.observationFlag = setTimeout(() => {
                    if (this.observing) {
                        this.refresh();
                        this.scheduleNextRefresh();
                    }
                }, this.observationFrequency * 1000);
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
            let node = this.node;
            let self = this;
            this.dataSource.refreshNodeDetails(node, function () {
                node.contentComponent = self;
                node.contentComponent.lastRefresh = new Date()
            });
        },

        beforeDestroy() {
            this.stopObserve();
            this.node.contentComponent = null;
        }
    }
</script>

<style lang="scss" scoped>
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
