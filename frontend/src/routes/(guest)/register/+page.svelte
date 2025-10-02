<script lang="ts">
import { onMount } from "svelte";
import { _ } from "svelte-i18n";
import "$lib/i18n";
import AuthForm from "$lib/components/molecules/AuthForm.svelte";
import type { ActionData, PageData } from "./$types";

export let form: ActionData;
export let data: PageData;

let registerKeyRequired = data.registerKeyRequired;

// コンポーネントマウント時に最新の設定を取得
onMount(async () => {
	try {
		const response = await fetch("/api/auth/registration-config");
		const config = await response.json();
		registerKeyRequired = config.registerKeyRequired;
	} catch (error) {
		console.error("Failed to fetch registration config:", error);
		// エラーの場合はデフォルト値を使用
		registerKeyRequired = false;
	}
});
</script>

<AuthForm
	title={$_('auth.register.title')}
	submitText={$_('auth.register.submit')}
	loadingText={$_('auth.register.submitting')}
	linkText={$_('auth.register.hasAccount')}
	linkHref="/login"
	showNameField={true}
	showRegisterKeyField={registerKeyRequired}
	error={form?.error}
/>