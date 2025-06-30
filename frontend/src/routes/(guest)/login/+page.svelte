<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";
import AuthForm from "$lib/components/molecules/AuthForm.svelte";
import type { ActionData } from "./$types";

export let form: ActionData;
</script>

<AuthForm
	title={$_('auth.login.title')}
	submitText={$_('auth.login.submit')}
	loadingText={$_('auth.login.submitting')}
	linkText={$_('auth.login.noAccount')}
	linkHref="/register"
	showNameField={false}
	error={form?.error}
/>