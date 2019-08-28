<template>
    <v-app>
        <div class="" id="app-container">
            <v-app-bar
                color="primary" dark>
                <v-toolbar-title class="app-logo" @click="refresh">Lagoon</v-toolbar-title>
                <div class="flex-grow-1"></div>
                <v-btn icon
                    @click="toGithub">
                    <font-awesome-icon :icon="['fab', 'github']" size="2x"/>
                </v-btn>
            </v-app-bar>
            <div id="content">
                <data-source 
                    @display-modal="showConfirmation">
                </data-source>
            </div>

            <v-snackbar
                v-model="showSnackbar"
                :timeout="4000">
                {{ snakebarText }}
                <v-btn
                    color="red darken-2"
                    text
                    dark
                    @click="showSnackbar = false">
                    Close
                </v-btn>
            </v-snackbar>
            <!-- https://www.npmjs.com/package/vue-js-modal -->
            <v-dialog/>
        </div>
    </v-app>
</template>

<script>
    import DataSource from './components/DataSource.vue'
    import EventBus from './eventBus'

    export default {
        name: 'app-container',

        components: {
            DataSource
        },

        data() {
            return {
                showSnackbar: false,
                snakebarText: ''
            }
        },

        methods: {
            showSnakebar: function(event) {
                this.snakebarText = event.message;
                this.showSnackbar = true;
            },

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

            toGithub: function () {
                window.open("https://github.com/ericjesse/lagoon", "_blank");
            },

            refresh: function () {
                window.document.location.reload();
            }
        },

        mounted() {
            EventBus.$on('display-snakebar', this.showSnakebar);
        }
    }
</script>

<style lang="scss">
    @import '../node_modules/splitpanes/dist/splitpanes.css';
    @import '../node_modules/bootstrap/scss/bootstrap.scss';
    @import '../node_modules/vuetify/dist/vuetify.min.css';
    @import '../node_modules/@mdi/font/css/materialdesignicons.css';
    @import "./assets/custom.scss";

    .app-logo {
        cursor: pointer;
    }
    #content {
        margin: 10px 10px 0 10px;
    }
</style>
