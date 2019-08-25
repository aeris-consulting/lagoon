<template>
    <div id="entrypoints">
        <div>
            <input id="overall-filter" v-model="dataSource.filter"/>
            <button @click="refresh()" class="btn btn-outline-primary">List</button>
            <span class="status" v-if="dataSource.status !== null">{{ dataSource.status }}</span>
        </div>

        <div :key="error.message" class="errors" v-for="error in dataSource.errors">{{ error.message }}</div>

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
                this.dataSource.status = 'Loading...';
                let self = this;
                this.root.clear();

                this.dataSource.listEntrypoints(null, 0, 0, receivedValues => {
                    self.dataSource.status = "Displaying...";
                    receivedValues.forEach(value => {
                        self.root.addChildNode(new Node(value.path, value.length, value.hasContent))
                    });
                    self.dataSource.status = null;
                }, error => {
                    self.dataSource.status = null;
                });
            }
        }
    }
</script>

<style lang="scss" scoped>
    input#overall-filter {
        margin-right: 15px;
    }

    div#entrypoints {
        text-align: left;
        margin-right: 10px;
    }

    div#entrypoints div.data {
        text-align: left;
    }
</style>
