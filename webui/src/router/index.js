import {createRouter, createWebHistory} from 'vue-router'
import {checkLiveness} from "../utils/livenessCheck.js";
import LoginView from '../views/LoginView.vue'
import HomeView from '../views/HomeView.vue'
import ManageView from '../views/ManageView.vue'
import ProfileView from '../views/ProfileView.vue'
import NotFoundView from '../views/404View.vue'
import ServiceUnavailableView from '../views/ServiceUnavailableView.vue'

const routes = [
	{
		path: '/',
		redirect: '/login'
	},
	{
		path: '/login',
		name: 'login',
		component: LoginView
	},
	{
		path: '/home',
		name: 'home',
		component: HomeView
	},
	{
		path: '/manage',
		name: 'manage',
		component: ManageView
	},
	{
		path: '/user/:profileUsername',
		name: 'profile',
		component: ProfileView,
		props: true
	},
	{
		path: '/service-unavailable',
		name: 'service-unavailable',
		component: ServiceUnavailableView
	},
	{
		path: '/:pathMatch(.*)*',
		name: '404',
		component: NotFoundView
	},
];

const router = createRouter({
	history: createWebHistory(import.meta.env.BASE_URL),
	routes: routes
})

router.beforeEach(async (to, from, next) => {
	if (to.path !== '/service-unavailable') {
		const isLive = await checkLiveness();
		if (isLive) {
			next();
		} else {
			next('/service-unavailable');
		}
	} else {
		next();
	}
});

export default router
