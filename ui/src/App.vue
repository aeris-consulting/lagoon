<template>
    <div class="container" id="app">
        <div class="container" id="header">
            <h1 @click="refresh" class="logo">Lagoon</h1>
        </div>
        <div class="container" id="content">
            <data-source @display-modal="showConfirmation"></data-source>
        </div>

        <!-- https://www.npmjs.com/package/vue-js-modal -->
        <v-dialog/>
    </div>
</template>

<script>
    import DataSource from './components/DataSource.vue'

    export default {
        name: 'app',

        components: {
            DataSource
        },

        data() {
            return {}
        },

        methods: {
            showConfirmation: function (event) {
                let buttons = [];
                if (event.noHandler) {
                    buttons.push({
                        title: 'Cancel',
                        default: true,
                        handler: () => {
                            this.$modal.hide('dialog');
                            event.noHandler();
                        }
                    });
                }
                if (event.yesHandler) {
                    buttons.push({
                        title: event.noHandler ? 'Confirm' : 'OK',
                        handler: () => {
                            this.$modal.hide('dialog');
                            event.yesHandler();
                        }
                    });
                }

                this.$modal.show('dialog', {
                    title: buttons.length > 1 ? 'Your confirmation is expected' : 'Message',
                    text: event.message,
                    buttons: buttons
                });
            },

            refresh: function () {
                window.document.location.reload();
            }
        }
    }
</script>

<style lang="scss">
    @import '../node_modules/splitpanes/dist/splitpanes.css';
    @import '../node_modules/bootstrap/scss/bootstrap.scss';
    @import "./assets/custom.scss";

    #header {
        h1.logo {
            cursor: pointer;
        }
    }
</style>
