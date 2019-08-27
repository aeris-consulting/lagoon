import Vue from 'vue'
import App from './App.vue'
import {library} from '@fortawesome/fontawesome-svg-core'
import {
    faAngleDown,
    faAngleRight,
    faClock,
    faCopy,
    faEdit,
    faEye,
    faPlay,
    faPlus,
    faStop,
    faSync,
    faTrash
} from '@fortawesome/free-solid-svg-icons'
import {
    faGithub
} from '@fortawesome/free-brands-svg-icons'
import {FontAwesomeIcon} from '@fortawesome/vue-fontawesome'
import VueClipboard from 'vue-clipboard2';
import VModal from 'vue-js-modal'
import vuetify from './plugins/vuetify'

library.add(faAngleRight, faAngleDown, faSync, faTrash, faEye, faClock, faEdit, faPlay, faStop, faCopy, faPlus, faGithub);
Vue.config.productionTip = false;
Vue.component('font-awesome-icon', FontAwesomeIcon);
Vue.use(VueClipboard);
VueClipboard.config.autoSetContainer = true;
Vue.use(VModal, {dialog: true})
// Vue.use(Vuetify)

new Vue({
    vuetify,
    render: h => h(App),
}).$mount('#app');
