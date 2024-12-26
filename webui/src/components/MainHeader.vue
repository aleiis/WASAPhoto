<script>

export default {
	methods: {
		pushToProfile() {
			const username = localStorage.getItem('username')
			const path = "/user/" + username
			this.$router.push(path)
		},
		onInputUpdate() {
			const bytes = new Blob([this.searchBar]).size;
			if (bytes > 16) {
				this.searchBar = this.searchBar.substring(0, this.searchBar.length - 1);
			}
		},
		logout() {
			localStorage.removeItem('userid');
			localStorage.removeItem('username');
			this.$router.push({path: '/login'});
		},
		async search() {
			if (this.searchBar) {
				const path = "/user/" + this.searchBar
				this.$router.push(path)
			}
		}
	},
	data() {
		return {
			searchBar: ''
		}
	}
}

</script>

<template>
	<header>
		<div class="header-logo">
			<img src="/src/assets/logo-naudit-retina-normal.png" style="width: auto; height: 40px;"/>
			<!--
			<h1>WASAPhoto</h1>
			-->
		</div>
		<div class="header-search">
			<input type="text" placeholder="Search" v-model="searchBar" @input="onInputUpdate" @keyup.enter="search"/>
			<button @click="search">Search</button>
		</div>
		<div class="header-menu">
			<img src="/src/assets/icons/media_manage_icon.png" @click="this.$router.push('/manage')"/>
			<img src="/src/assets/icons/home_icon.png" @click="this.$router.push('/home')"/>
			<img src="/src/assets/icons/account_circle_icon.png" @click="pushToProfile"/>
			<img src="/src/assets/icons/logout_icon.png" @click="logout">
		</div>
	</header>
</template>

<style>

header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 15px;
	background-color: #e6e6e6;
	border-bottom: 1px solid #bebebe;
	margin-bottom: 0px;
	height: 50px;
}

.header-logo {
	display: flex;
	align-items: center;
	margin-left: 20px;
}

.header-logo h1 {
	font-family: "PoetsenOne";
	font-size: 18px;
	margin-left: 20px;
	color: #008080;
}

.header-search {
	display: flex;
	align-items: center;
}

.header-search input[type="text"] {
	width: 300px;
	height: 100%;
	padding: 8px;
	border: 1px solid #ccc;
	border-radius: 4px;
	font-size: 14px;
}

.header-search button {
	padding: 8px 16px;
	border: 1px solid #ccc;
	border-radius: 4px;
	font-size: 14px;
	margin-left: 8px;
	cursor: pointer;
}

.header-menu img {
	width: 30px;
	height: 30px;
	margin-right: 20px;
	cursor: pointer;
}

</style>
