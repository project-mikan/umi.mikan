<script lang="ts">
  import { _ } from "svelte-i18n";
  import "$lib/i18n";
  import { enhance } from "$app/forms";
  import { page } from "$app/stores";
  import Alert from "$lib/components/atoms/Alert.svelte";
  import Button from "$lib/components/atoms/Button.svelte";
  import Link from "$lib/components/atoms/Link.svelte";
  import type { ActionData, PageData } from "./$types";

  export let data: PageData;
  export let form: ActionData;

  let submitting = false;

  $: csrfToken = $page.data.csrfToken as string | undefined;

  // 同意画面に接続先ホストを表示し、ユーザーが何に同意しているか確認できるようにする。
  // redirect_uriはバックエンド側で登録済みURLとの一致検証済みだが、URLとして不正な
  // 値が渡ってきた場合に備えてパース失敗時はredirect_uriの生値をフォールバック表示する。
  $: redirectHost = (() => {
    if (!data.redirectUri) return "";
    try {
      return new URL(data.redirectUri).host;
    } catch {
      return data.redirectUri;
    }
  })();

  // action成功時にauthorization codeを含むredirect_uriへ遷移する。
  // SvelteKitのuse:enhanceはactionの戻り値をformに反映するだけで自動遷移はしないため、
  // ここで明示的にwindow.location.hrefを更新してMCPクライアント側に戻す。
  $: if (form?.success && form.redirectUrl) {
    window.location.href = form.redirectUrl;
  }

  // MCPサーバー（別オリジン）からのリダイレクトチェーン経由の初回アクセスでは
  // SameSite=StrictのログインCookieが送信されず未ログイン判定になるため、
  // 同一オリジンへの自己リダイレクトを一度だけ行い「同一サイトナビゲーション」として
  // 再読み込みさせる（retry=1を付与し、無限ループを防ぐ）。
  $: if (data.needsSameSiteRetry) {
    const retryUrl = new URL($page.url);
    retryUrl.searchParams.set("retry", "1");
    window.location.replace(retryUrl.toString());
  }
</script>

<svelte:head>
  <title>{$_("mcpOAuth.title")}</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 px-4">
	<div class="max-w-md w-full space-y-8">
		<div>
			<div class="flex justify-center mb-4">
				<img src="/favicon.png" alt="umi.mikan" class="h-20 w-20" />
			</div>
			<p class="text-center text-2xl font-bold text-gray-700 dark:text-gray-300 mb-2">{$_('common.appName')}</p>
			<h2 class="text-center text-xl font-semibold text-gray-900 dark:text-white">
				{$_("mcpOAuth.title")}
			</h2>
		</div>

		<div class="mt-8 space-y-6">
			{#if data.needsSameSiteRetry}
				<p class="text-center text-gray-700 dark:text-gray-300">
					{$_("mcpOAuth.checkingLogin")}
				</p>
			{:else if data.invalidRequest}
				<Alert type="error">
					{$_("mcpOAuth.invalidRequest")}
				</Alert>
			{:else if !data.isAuthenticated}
				<p class="text-center text-gray-700 dark:text-gray-300">
					{$_("mcpOAuth.loginRequired")}
				</p>
				<Link href="/login" variant="primary">
					{$_("mcpOAuth.goToLogin")}
				</Link>
			{:else}
				<p class="text-center text-gray-700 dark:text-gray-300">
					{$_("mcpOAuth.description")}
				</p>
				{#if redirectHost}
					<p class="text-center text-sm text-gray-500 dark:text-gray-400">
						{$_("mcpOAuth.redirectTo", { values: { host: redirectHost } })}
					</p>
				{/if}

				{#if form?.error}
					<Alert type="error">
						{$_(`mcpOAuth.errors.${form.error}`)}
					</Alert>
				{/if}

				<form
					method="POST"
					action="?/consent"
					use:enhance={() => {
						submitting = true;
						return async ({ update }) => {
							try {
								await update();
							} finally {
								submitting = false;
							}
						};
					}}
				>
					<input type="hidden" name="csrfToken" value={csrfToken} />
					<input type="hidden" name="client_id" value={data.clientId} />
					<input type="hidden" name="redirect_uri" value={data.redirectUri} />
					<input
						type="hidden"
						name="code_challenge"
						value={data.codeChallenge}
					/>
					<input
						type="hidden"
						name="code_challenge_method"
						value={data.codeChallengeMethod}
					/>
					<input type="hidden" name="state" value={data.state} />

					<Button
						type="submit"
						variant="primary"
						size="md"
						disabled={submitting}
						class="w-full flex justify-center"
					>
						{submitting
							? $_("mcpOAuth.authorizing")
							: $_("mcpOAuth.authorize")}
					</Button>
				</form>
			{/if}
		</div>
	</div>
</div>
