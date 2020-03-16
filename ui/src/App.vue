<template>
    <v-app>
        <div class="" id="app-container">
            <div class="alerts-container">    
                <v-alert
                    :key="i" class="errors" v-for="(error, i) in errors"
                    border="left"
                    type="error"
                    dark>
                    {{ error.message }}
                    <template v-slot:append>
                        <v-btn
                            @click="dismissErrorMessage(i)"
                            class="mx-2" icon>
                            <v-icon dark>mdi-close</v-icon>
                        </v-btn>
                    </template>
                </v-alert>
            </div>
            <v-app-bar
                color="primary" dark>
                <v-toolbar-title @click="refresh" class="app-logo">Lagoon
                    <span class="datasource-name" v-if="selectedDatasource != null">{{ selectedDatasource.name }}<template
                            v-if="selectedDatasource.readonly">&nbsp;(Read-only)</template></span>
                </v-toolbar-title>
                <div class="flex-grow-1"></div>
                <v-btn icon
                       @click="openTerminal"
                       v-if="selectedDatasourceId">
                    <font-awesome-icon :icon="['fa', 'terminal']"/>
                </v-btn>
                <v-btn icon
                       @click="toGithub">
                    <font-awesome-icon :icon="['fab', 'github']" size="2x"/>
                </v-btn>
            </v-app-bar>
            <div id="content">
                <data-source
                        ref="dataSource">
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
    import { mapState } from 'vuex'
    import { DISSMISS_ERROR } from './store/actions.type'

    export default {
        name: 'app-container',

        components: {
            DataSource
        },

        computed: mapState({
            selectedDatasourceId: state => state.datasource.selectedDatasourceId,
            errors: state => state.error.errors,
        }),

        data() {
            return {
                selectedDatasource: null,
                showSnackbar: false,
                snakebarText: '',
                showTerminalButton: false,
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

            openTerminal: function () {
                EventBus.$emit('open-terminal');
            },

            toGithub: function () {
                window.open("https://github.com/ericjesse/lagoon", "_blank");
            },

            refresh: function () {
                window.document.location.reload();
            },

            dismissErrorMessage: function(errorIndex) {
                this.$store.dispatch(DISSMISS_ERROR, errorIndex)
            }
        },

        mounted() {
            EventBus.$on('display-modal', this.showConfirmation)
            EventBus.$on('display-snakebar', this.showSnakebar);
            EventBus.$on('datasource-set', (event) => {
                // eslint-disable-next-line
                console.log(event.datasource);
                this.selectedDatasource = event.datasource;
                document.title = 'Lagoon - ' + event.datasource.name;
                this.showTerminalButton = true;
            });
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

    .datasource-name {
        margin-left: 20px;
        font-size: 14px;
    }

    .alerts-container {
        position: fixed;
        z-index: 9999;
        top: 80px;
        right: 20px;
    }
</style>
