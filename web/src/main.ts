import { createApp } from "vue";
import router from "@/app/router";
import App from "@/app/App.vue";
import "@/app/styles/index.css";

createApp(App).use(router).mount("#app");
