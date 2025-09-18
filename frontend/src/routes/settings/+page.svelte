<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";
import { enhance } from "$app/forms";
import Modal from "$lib/components/molecules/Modal.svelte";
import type { ActionData, PageData } from "./$types";

export let form: ActionData;
export let data: PageData;

let usernameLoading = false;
let passwordLoading = false;
let llmTokenLoading = false;
let deleteLLMKeyLoading = false;
let deleteAccountLoading = false;

// Modal states
let showDeleteLLMTokenConfirm = false;
let showDeleteAccountConfirm = false;

// Password visibility toggles
let showCurrentPassword = false;
let showNewPassword = false;
let showConfirmPassword = false;

// Get existing LLM key for Gemini (provider 0)
$: existingLLMToken =
	data.user?.llmKeys?.find((key) => key.llmProvider === 0)?.key || "";

// Modal helper functions
function confirmDeleteLLMToken() {
	showDeleteLLMTokenConfirm = true;
}

function cancelDeleteLLMToken() {
	showDeleteLLMTokenConfirm = false;
}

function handleDeleteLLMToken() {
	showDeleteLLMTokenConfirm = false;
	// Submit the delete form
	const form = document.createElement("form");
	form.method = "POST";
	form.action = "?/deleteLLMKey";

	const input = document.createElement("input");
	input.type = "hidden";
	input.name = "llmProvider";
	input.value = "0";
	form.appendChild(input);

	document.body.appendChild(form);
	deleteLLMKeyLoading = true;
	form.submit();
}

function confirmDeleteAccount() {
	showDeleteAccountConfirm = true;
}

function cancelDeleteAccount() {
	showDeleteAccountConfirm = false;
}

function handleDeleteAccount() {
	showDeleteAccountConfirm = false;
	// Submit the delete form
	const form = document.createElement("form");
	form.method = "POST";
	form.action = "?/deleteAccount";

	document.body.appendChild(form);
	deleteAccountLoading = true;
	form.submit();
}
</script>

<svelte:head>
	<title>{$_("settings.title")}</title>
</svelte:head>

<main class="container mx-auto px-4 py-8">
	<h1 class="text-3xl font-bold mb-8">{$_("settings.title")}</h1>

	<div class="max-w-2xl mx-auto space-y-8">
		<!-- エラー/成功メッセージ -->
		{#if form?.error}
			<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
				{$_(`settings.messages.${form.error}`) || form.error}
			</div>
		{/if}
		{#if form?.success}
			<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">
				{$_(`settings.messages.${form.message}`) || form.message}
			</div>
		{/if}

		<!-- ユーザー名変更セクション -->
		<section class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
			<h2 class="text-xl font-semibold mb-4">{$_("settings.username.title")}</h2>
			<form
				method="POST"
				action="?/updateUsername"
				class="space-y-4"
				use:enhance={() => {
					usernameLoading = true;
					return async ({ update }) => {
						usernameLoading = false;
						await update();
					};
				}}
			>
				<div>
					<label for="username" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
						{$_("settings.username.label")}
					</label>
					<input
						type="text"
						id="username"
						name="username"
						required
						maxlength="20"
						disabled={usernameLoading}
						value={data.user?.name || ""}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100"
						placeholder={$_("settings.username.placeholder")}
					/>
				</div>
				<button
					type="submit"
					disabled={usernameLoading}
					class="bg-blue-500 hover:bg-blue-600 disabled:bg-blue-300 text-white font-medium py-2 px-4 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
				>
					{usernameLoading ? $_("common.loading") : $_("settings.username.save")}
				</button>
			</form>
		</section>

		<!-- パスワード変更セクション -->
		<section class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
			<h2 class="text-xl font-semibold mb-4">{$_("settings.password.title")}</h2>
			<form
				method="POST"
				action="?/changePassword"
				class="space-y-4"
				use:enhance={() => {
					passwordLoading = true;
					return async ({ update }) => {
						passwordLoading = false;
						await update();
					};
				}}
			>
				<div>
					<label for="currentPassword" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
						{$_("settings.password.currentLabel")}
					</label>
					<div class="relative">
						<input
							type={showCurrentPassword ? "text" : "password"}
							id="currentPassword"
							name="currentPassword"
							required
							disabled={passwordLoading}
							class="w-full px-3 py-2 pr-10 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100"
							placeholder={$_("settings.password.currentPlaceholder")}
						/>
						<button
							type="button"
							class="absolute inset-y-0 right-0 px-3 flex items-center text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300"
							on:click={() => showCurrentPassword = !showCurrentPassword}
							disabled={passwordLoading}
						>
							{#if showCurrentPassword}
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L3 3m6.878 6.878L21 21"></path>
								</svg>
							{:else}
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
								</svg>
							{/if}
						</button>
					</div>
				</div>
				<div>
					<label for="newPassword" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
						{$_("settings.password.newLabel")}
					</label>
					<div class="relative">
						<input
							type={showNewPassword ? "text" : "password"}
							id="newPassword"
							name="newPassword"
							required
							minlength="8"
							disabled={passwordLoading}
							class="w-full px-3 py-2 pr-10 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100"
							placeholder={$_("settings.password.newPlaceholder")}
						/>
						<button
							type="button"
							class="absolute inset-y-0 right-0 px-3 flex items-center text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300"
							on:click={() => showNewPassword = !showNewPassword}
							disabled={passwordLoading}
						>
							{#if showNewPassword}
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L3 3m6.878 6.878L21 21"></path>
								</svg>
							{:else}
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
								</svg>
							{/if}
						</button>
					</div>
				</div>
				<div>
					<label for="confirmPassword" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
						{$_("settings.password.confirmLabel")}
					</label>
					<div class="relative">
						<input
							type={showConfirmPassword ? "text" : "password"}
							id="confirmPassword"
							name="confirmPassword"
							required
							minlength="8"
							disabled={passwordLoading}
							class="w-full px-3 py-2 pr-10 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100"
							placeholder={$_("settings.password.confirmPlaceholder")}
						/>
						<button
							type="button"
							class="absolute inset-y-0 right-0 px-3 flex items-center text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300"
							on:click={() => showConfirmPassword = !showConfirmPassword}
							disabled={passwordLoading}
						>
							{#if showConfirmPassword}
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 711.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L3 3m6.878 6.878L21 21"></path>
								</svg>
							{:else}
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
								</svg>
							{/if}
						</button>
					</div>
				</div>
				<button
					type="submit"
					disabled={passwordLoading}
					class="bg-blue-500 hover:bg-blue-600 disabled:bg-blue-300 text-white font-medium py-2 px-4 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
				>
					{passwordLoading ? $_("common.loading") : $_("settings.password.save")}
				</button>
			</form>
		</section>

		<!-- LLMトークン変更セクション -->
		<section class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
			<h2 class="text-xl font-semibold mb-4">{$_("settings.llmToken.title")}</h2>
			<form
				method="POST"
				action="?/updateLLMKey"
				class="space-y-4"
				use:enhance={() => {
					llmTokenLoading = true;
					return async ({ update }) => {
						llmTokenLoading = false;
						await update();
					};
				}}
			>
				<div>
					<label for="llmProvider" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
						{$_("settings.llmToken.providerLabel")}
					</label>
					<select
						id="llmProvider"
						name="llmProvider"
						disabled={llmTokenLoading}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100"
					>
						<option value="0">{$_("settings.llmToken.provider.gemini")}</option>
					</select>
				</div>
				<div>
					<label for="llmToken" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
						{$_("settings.llmToken.tokenLabel")}
					</label>
					<input
						type="text"
						id="llmToken"
						name="llmKey"
						required
						maxlength="100"
						disabled={llmTokenLoading}
						value={existingLLMToken}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100"
						placeholder={$_("settings.llmToken.tokenPlaceholder")}
					/>
				</div>
				<button
					type="submit"
					disabled={llmTokenLoading}
					class="bg-blue-500 hover:bg-blue-600 disabled:bg-blue-300 text-white font-medium py-2 px-4 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
				>
					{llmTokenLoading ? $_("common.loading") : $_("settings.llmToken.save")}
				</button>
			</form>

			<!-- LLM Token Delete Section -->
			{#if existingLLMToken}
				<div class="mt-4">
					<button
						type="button"
						disabled={deleteLLMKeyLoading}
						on:click={confirmDeleteLLMToken}
						class="bg-red-500 hover:bg-red-600 disabled:bg-red-300 text-white font-medium py-2 px-4 rounded-md focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
					>
						{deleteLLMKeyLoading ? $_("common.loading") : $_("settings.deleteToken.button")}
					</button>
				</div>
			{/if}
		</section>

		<!-- Account Deletion Section -->
		<hr class="my-8 border-gray-300 dark:border-gray-600" />

		<section class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
			<h2 class="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
				{$_("settings.deleteAccount.title")}
			</h2>
			<div class="mb-4">
				<p class="text-gray-700 dark:text-gray-300 mb-2">
					{$_("settings.deleteAccount.warning")}
				</p>
			</div>
			<button
				type="button"
				disabled={deleteAccountLoading}
				on:click={confirmDeleteAccount}
				class="bg-red-600 hover:bg-red-700 disabled:bg-red-300 text-white font-medium py-2 px-4 rounded-md focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
			>
				{deleteAccountLoading ? $_("common.loading") : $_("settings.deleteAccount.button")}
			</button>
		</section>
	</div>
</main>

<!-- LLM Token Delete Confirmation Modal -->
<Modal
	isOpen={showDeleteLLMTokenConfirm}
	title={$_("settings.deleteToken.confirm")}
	confirmText={$_("settings.deleteToken.button")}
	cancelText={$_("diary.cancel")}
	variant="danger"
	onConfirm={handleDeleteLLMToken}
	onCancel={cancelDeleteLLMToken}
>
	<p class="text-sm text-gray-500 dark:text-gray-400">
		{$_("settings.deleteToken.confirmMessage")}
	</p>
</Modal>

<!-- Account Delete Confirmation Modal -->
<Modal
	isOpen={showDeleteAccountConfirm}
	title={$_("settings.deleteAccount.confirm")}
	confirmText={$_("settings.deleteAccount.button")}
	cancelText={$_("diary.cancel")}
	variant="danger"
	onConfirm={handleDeleteAccount}
	onCancel={cancelDeleteAccount}
>
	<p class="text-sm text-gray-500 dark:text-gray-400 mb-2">
		{$_("settings.deleteAccount.confirmMessage")}
	</p>
	<p class="text-sm text-red-600 dark:text-red-400 font-medium">
		{$_("settings.deleteAccount.warning")}
	</p>
</Modal>