<script lang="ts">
  import { _ } from "svelte-i18n";
  import "$lib/i18n";
  import { enhance } from "$app/forms";
  import { page } from "$app/stores";
  import type { ActionData, PageData } from "./$types";

  export let data: PageData;
  export let form: ActionData;

  let submitting = false;

  $: csrfToken = $page.data.csrfToken as string | undefined;

  // action成功時にauthorization codeを含むredirect_uriへ遷移する。
  // SvelteKitのuse:enhanceはactionの戻り値をformに反映するだけで自動遷移はしないため、
  // ここで明示的にwindow.location.hrefを更新してMCPクライアント側に戻す。
  $: if (form?.success && form.redirectUrl) {
    window.location.href = form.redirectUrl;
  }
</script>

<svelte:head>
  <title>{$_("mcpOAuth.title")}</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-gray-50 dark:bg-gray-900 px-4">
  <div class="w-full max-w-md rounded-lg bg-white dark:bg-gray-800 p-8 shadow">
    <h1 class="mb-4 text-xl font-bold text-gray-900 dark:text-white">
      {$_("mcpOAuth.title")}
    </h1>

    {#if data.invalidRequest}
      <p class="text-red-600 dark:text-red-400">
        {$_("mcpOAuth.invalidRequest")}
      </p>
    {:else if !data.isAuthenticated}
      <p class="mb-4 text-gray-700 dark:text-gray-300">
        {$_("mcpOAuth.loginRequired")}
      </p>
      <a
        href="/login"
        class="inline-block rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
      >
        {$_("mcpOAuth.goToLogin")}
      </a>
    {:else}
      <p class="mb-6 text-gray-700 dark:text-gray-300">
        {$_("mcpOAuth.description")}
      </p>

      {#if form?.error}
        <p class="mb-4 text-red-600 dark:text-red-400">
          {$_(`mcpOAuth.errors.${form.error}`)}
        </p>
      {/if}

      <form
        method="POST"
        action="?/consent"
        use:enhance={() => {
          submitting = true;
          return async ({ update }) => {
            await update();
            submitting = false;
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

        <button
          type="submit"
          disabled={submitting}
          class="w-full rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
        >
          {submitting
            ? $_("mcpOAuth.authorizing")
            : $_("mcpOAuth.authorize")}
        </button>
      </form>
    {/if}
  </div>
</div>
