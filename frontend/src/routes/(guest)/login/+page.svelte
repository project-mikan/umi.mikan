<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";
import AuthForm from "$lib/components/molecules/AuthForm.svelte";
import type { ActionData } from "./$types";

// フォームデータに含まれる可能性のあるプロパティを定義
interface FormDataWithRateLimit {
	error?: string;
	isRateLimited?: boolean;
}

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
	isRateLimited={(form as FormDataWithRateLimit)?.isRateLimited ?? false}
/>