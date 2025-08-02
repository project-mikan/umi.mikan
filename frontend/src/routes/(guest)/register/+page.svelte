<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";
import AuthForm from "$lib/components/molecules/AuthForm.svelte";
import type { ActionData } from "./$types";

export let form: ActionData;
</script>

<AuthForm
	title={$_('auth.register.title')}
	submitText={$_('auth.register.submit')}
	loadingText={$_('auth.register.submitting')}
	linkText={$_('auth.register.hasAccount')}
	linkHref="/login"
	showNameField={true}
	error={form?.error}
/>