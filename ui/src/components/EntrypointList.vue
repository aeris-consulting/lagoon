<template>
    <div id="entrypoints">
        <div class="alerts-container">
            <v-alert
                :key="i" class="errors" v-for="(error, i) in dataSource.errors"
                :value="true"
                @input="dismissErrorMessage(i)"
                border="left"
                close-text="Close Alert"
                type="error"
                dark
                dismissible>
                {{ error.message }}
            </v-alert>
        </div>
        <div>
            <div>
                <div class="filter-container">
                    <v-text-field
                        v-model="dataSource.filter"
                        label="Filter"
                    ></v-text-field>
                </div>
                <v-btn class="" color="primary" @click="refresh()">List</v-btn>
                <v-progress-circular
                    class="loading-circle"
                    v-if="dataSource.status !== null"
                    indeterminate
                    color="green">
                </v-progress-circular>
            </div>
            <!-- <span class="status" v-if="dataSource.status !== null">{{ dataSource.status }}</span> -->
        </div>

        <div v-if="root.hasChildren() && root.children !== null">
            <entrypoint-children @display-modal="showConfirmation"
                                 v-bind:children="root.children.values()"
                                 v-bind:dataSource="dataSource"></entrypoint-children>
        </div>
    </div>
</template>

<script>
    import EntrypointChildren from "./EntrypointChildren";
    import Node from "../models/Node";

    export default {
        name: 'EntrypointList',
        components: {EntrypointChildren},

        props: {
            dataSource: Object,
        },

        data() {
            return {
                root: new Node('', 0)
            }
        },

        methods: {
            showConfirmation: function (event) {
                this.$emit('display-modal', event);
            },

            refresh: function () {
                this.dataSource.status = 'loading';
                let self = this;
                this.root.clear();
                this.dataSource.listEntrypoints(null, 0, 0, receivedValues => {
                    receivedValues.forEach(value => {
                        self.root.addChildNode(new Node(value.path, value.length, value.hasContent))
                    });
                    self.dataSource.status = null;
                }, () => {
                    self.dataSource.status = null;
                });
            },

            dismissErrorMessage: function(errorIndex) {
                this.dataSource.errors.splice(errorIndex, 1);
            }
        }
    }
</script>

<style lang="scss" scoped>
    .filter-container {
        margin-right: 15px;
        width: 200px;
        display: inline-block;
    }

    .alerts-container {
        position: fixed;
        z-index: 9999;
        top: 80px;
        right: 20px;
    }

    div#entrypoints {
        text-align: left;
        margin-right: 10px;
    }

    div#entrypoints div.data {
        text-align: left;
    }

    .loading-circle {
        margin-left: 10px;
    }
</style>
