<script lang="ts">
  import { onMount, untrack } from "svelte";
  import { _ } from "svelte-i18n";
  import "#lib/i18n.js";
  import AuthForm from "#lib/components/molecules/AuthForm.svelte";
  import type { ActionData, PageData } from "./$types";

  interface Props {
    form: ActionData;
    data: PageData;
  }

  let { form, data }: Props = $props();

  // 初期値だけ SSR の値を使い、マウント後に最新の設定で上書きする
  let registerKeyRequired = $state(untrack(() => data.registerKeyRequired));

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
