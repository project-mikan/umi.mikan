<script lang="ts">
import "$lib/i18n";
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