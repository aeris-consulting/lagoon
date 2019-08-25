<template>
    <div class="data" v-if="values.length" v-show="visible">
        <entrypoint @display-modal="showConfirmation" v-bind:dataSource="dataSource" v-bind:key="child.name"
                    v-bind:node="child" v-for="child in values"></entrypoint>
    </div>
</template>

<script>
    import Entrypoint from "./Entrypoint";

    export default {
        name: 'entrypoint-children',

        components: {Entrypoint},

        props: ['children', 'dataSource'],

        data() {
            return {
                visible: true,
            }
        },

        methods: {
            showConfirmation: function (event) {
                this.$emit('display-modal', event);
            },

            select: function (event) {
                this.$emit('selectedEntrypoint', event);
            }
        },

        computed: {
            values() {
                let values = [];
                let v;
                for (v of this.children) {
                    values.push(v);
                }
                values.sort();
                return values;
            }
        }
    }
</script>

<style lang="scss" scoped>

</style>
