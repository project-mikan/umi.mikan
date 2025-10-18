<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";
import { enhance } from "$app/forms";
import { onMount } from "svelte";
import Modal from "$lib/components/molecules/Modal.svelte";
import SettingsNav from "$lib/components/molecules/SettingsNav.svelte";
import type { ActionData, PageData } from "./$types";

export let form: ActionData;
export let data: PageData;

// Helper function to check if message belongs to specific action
function isMessageForAction(actionName: string): boolean {
	return form?.action === actionName;
}

let usernameLoading = false;
let passwordLoading = false;
let llmTokenLoading = false;
let autoSummaryLoading = false;
let deleteLLMKeyLoading = false;
let deleteAccountLoading = false;

// Modal states
let showDeleteLLMTokenConfirm = false;
let showDeleteAccountConfirm = false;

// Password visibility toggles
let showCurrentPassword = false;
let showNewPassword = false;
let showConfirmPassword = false;

// Get existing LLM key for Gemini (provider 1)
$: existingLLMKey = data.user?.llmKeys?.find((key) => key.llmProvider === 1);
$: existingLLMToken = existingLLMKey?.key || "";

// Local state for checkbox values
let autoSummaryDaily = false;
let autoSummaryMonthly = false;
let autoLatestTrend = false;

// Update local state when data changes
$: {
	if (existingLLMKey) {
		autoSummaryDaily = existingLLMKey.autoSummaryDaily || false;
		autoSummaryMonthly = existingLLMKey.autoSummaryMonthly || false;
		autoLatestTrend = existingLLMKey.autoLatestTrendEnabled || false;
	}
}

// Active section tracking
let activeSection = "";

// Mobile navigation state
let isMobileNavOpen = false;

function toggleMobileNav() {
	isMobileNavOpen = !isMobileNavOpen;
}

// Intersection Observer for tracking active section
onMount(() => {
	let isUserScrolling = false;
	let scrollTimeout: ReturnType<typeof setTimeout>;

	// Listen for scroll events to detect user scrolling
	const handleScroll = () => {
		isUserScrolling = true;
		if (scrollTimeout) {
			clearTimeout(scrollTimeout);
		}
		scrollTimeout = setTimeout(() => {
			isUserScrolling = false;
		}, 150);
	};

	window.addEventListener("scroll", handleScroll);

	const observer = new IntersectionObserver(
		(entries) => {
			// Only update if not currently scrolling programmatically
			if (!isUserScrolling) {
				entries.forEach((entry) => {
					if (entry.isIntersecting) {
						activeSection = entry.target.id;
					}
				});
			}
		},
		{ threshold: 0.3, rootMargin: "-20% 0px -60% 0px" },
	);

	const sections = document.querySelectorAll("section[id]");
	for (const section of sections) {
		observer.observe(section);
	}

	return () => {
		window.removeEventListener("scroll", handleScroll);
		for (const section of sections) {
			observer.unobserve(section);
		}
		if (scrollTimeout) {
			clearTimeout(scrollTimeout);
		}
	};
});

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
	input.value = "1";
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
	<!-- Mobile Header with Hamburger Menu -->
	<div class="md:hidden flex items-center mb-6">
		<button
			type="button"
			on:click={toggleMobileNav}
			class="p-2 mr-4 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg shadow-sm hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
			aria-label={$_("settings.nav.title")}
		>
			<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path>
			</svg>
		</button>
		<h1 class="text-2xl font-bold">{$_("settings.title")}</h1>
	</div>

	<!-- Desktop Title -->
	<h1 class="hidden md:block text-3xl font-bold mb-8">{$_("settings.title")}</h1>

	<div class="flex gap-8 max-w-6xl mx-auto">
		<!-- Settings Navigation -->
		<aside class="w-64 flex-shrink-0 hidden md:block">
			<SettingsNav {activeSection} isOpen={false} onToggle={() => {}} />
		</aside>

		<!-- Mobile Settings Navigation (only rendered on mobile) -->
		<div class="md:hidden">
			<SettingsNav {activeSection} isOpen={isMobileNavOpen} onToggle={toggleMobileNav} />
		</div>

		<!-- Settings Content -->
		<div class="flex-1 space-y-8 w-full md:w-auto">

			<!-- ユーザー設定セクション -->
			<div class="space-y-8">
				<div>
					<h2 class="text-2xl font-bold mb-6 text-gray-900 dark:text-white">
						{$_("settings.nav.userSettings")}
					</h2>

					<!-- ユーザー名変更セクション -->
					<section id="username" class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 mb-6">
						<h3 class="text-xl font-semibold mb-4">{$_("settings.username.title")}</h3>
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
				<!-- ユーザー名変更メッセージ -->
				{#if form?.error && isMessageForAction("updateUsername")}
					<div class="mt-3 bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded auto-phrase-target">
						{$_(`settings.messages.${form.error}`) || form.error}
					</div>
				{/if}
				{#if form?.success && isMessageForAction("updateUsername")}
					<div class="mt-3 bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded auto-phrase-target">
						{$_(`settings.messages.${form.message}`) || form.message}
					</div>
				{/if}
			</form>
					</section>

					<!-- パスワード変更セクション -->
					<section id="password" class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
						<h3 class="text-xl font-semibold mb-4">{$_("settings.password.title")}</h3>
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
				<!-- パスワード変更メッセージ -->
				{#if form?.error && isMessageForAction("changePassword")}
					<div class="mt-3 bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded auto-phrase-target">
						{$_(`settings.messages.${form.error}`) || form.error}
					</div>
				{/if}
				{#if form?.success && isMessageForAction("changePassword")}
					<div class="mt-3 bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded auto-phrase-target">
						{$_(`settings.messages.${form.message}`) || form.message}
					</div>
				{/if}
			</form>
					</section>
				</div>
			</div>

			<!-- LLM設定セクション -->
			<div class="space-y-8">
				<div>
					<h2 class="text-2xl font-bold mb-6 text-gray-900 dark:text-white">
						{$_("settings.nav.llmSettings")}
					</h2>

					<!-- LLMトークン変更セクション -->
					<section id="llm-token" class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 mb-6">
						<h3 class="text-xl font-semibold mb-4">{$_("settings.llmToken.title")}</h3>
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
						<option value="1">{$_("settings.llmToken.provider.gemini")}</option>
					</select>
				</div>
				<div>
					<label for="llmToken" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
						{$_("settings.llmToken.tokenLabel")}
					</label>
					<p class="text-xs text-gray-500 dark:text-gray-400 mb-2 auto-phrase-target">
						{$_("settings.llmToken.tokenHelp")} <a href="https://aistudio.google.com/apikey" target="_blank" rel="noopener noreferrer" class="text-blue-500 hover:text-blue-600 underline">https://aistudio.google.com/apikey</a>
					</p>
					<p class="text-xs text-orange-600 dark:text-orange-400 mb-2 bg-orange-50 dark:bg-orange-900/20 p-2 rounded border border-orange-200 dark:border-orange-800 auto-phrase-target">
						{$_("settings.llmToken.freeWarning")}
					</p>
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
				<!-- LLMトークン変更メッセージ -->
				{#if form?.error && isMessageForAction("updateLLMKey")}
					<div class="mt-3 bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded auto-phrase-target">
						{$_(`settings.messages.${form.error}`) || form.error}
					</div>
				{/if}
				{#if form?.success && isMessageForAction("updateLLMKey")}
					<div class="mt-3 bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded auto-phrase-target">
						{$_(`settings.messages.${form.message}`) || form.message}
					</div>
				{/if}
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
					<!-- LLMトークン削除メッセージ -->
					{#if form?.error && isMessageForAction("deleteLLMKey")}
						<div class="mt-3 bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded auto-phrase-target">
							{$_(`settings.messages.${form.error}`) || form.error}
						</div>
					{/if}
					{#if form?.success && isMessageForAction("deleteLLMKey")}
						<div class="mt-3 bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded auto-phrase-target">
							{$_(`settings.messages.${form.message}`) || form.message}
						</div>
					{/if}
				</div>
			{/if}
					</section>

					<!-- 自動要約設定セクション -->
					{#if existingLLMToken}
						<section id="auto-summary" class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 mb-6">
							<h3 class="text-xl font-semibold mb-4">{$_("settings.autoSummary.title")}</h3>
				<p class="text-sm text-gray-600 dark:text-gray-400 mb-4 auto-phrase-target">
					{$_("settings.autoSummary.description")}
				</p>
				<form
					method="POST"
					action="?/updateAutoSummarySettings"
					class="space-y-4"
					use:enhance={() => {
						autoSummaryLoading = true;
						return async ({ result, update }) => {
							autoSummaryLoading = false;
							// If the action returned updated user data, use it
							if (result.type === 'success' && result.data?.user) {
								data = { ...data, user: result.data.user as typeof data.user };
							}
							await update({ reset: false });
						};
					}}
				>
					<input type="hidden" name="llmProvider" value="1" />

					<div class="space-y-3">
						<label class="flex items-center">
							<input
								type="checkbox"
								name="autoSummaryDaily"
								bind:checked={autoSummaryDaily}
								disabled={autoSummaryLoading}
								class="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50 disabled:bg-gray-100"
							/>
							<span class="ml-2 text-sm text-gray-700 dark:text-gray-300">
								{$_("settings.autoSummary.dailyLabel")}
							</span>
						</label>

						<label class="flex items-center">
							<input
								type="checkbox"
								name="autoSummaryMonthly"
								bind:checked={autoSummaryMonthly}
								disabled={autoSummaryLoading}
								class="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50 disabled:bg-gray-100"
							/>
							<span class="ml-2 text-sm text-gray-700 dark:text-gray-300">
								{$_("settings.autoSummary.monthlyLabel")}
							</span>
						</label>

						<label class="flex items-center">
							<input
								type="checkbox"
								name="autoLatestTrendEnabled"
								bind:checked={autoLatestTrend}
								disabled={autoSummaryLoading}
								class="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50 disabled:bg-gray-100"
							/>
							<span class="ml-2 text-sm text-gray-700 dark:text-gray-300">
								{$_("latestTrend.settings.autoGenerate")}
							</span>
						</label>
						<p class="text-xs text-gray-500 dark:text-gray-400 ml-6 auto-phrase-target">
							{$_("latestTrend.settings.autoGenerateDescription")}
						</p>
					</div>

					<button
						type="submit"
						disabled={autoSummaryLoading}
						class="bg-blue-500 hover:bg-blue-600 disabled:bg-blue-300 text-white font-medium py-2 px-4 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
					>
						{autoSummaryLoading ? $_("common.loading") : $_("settings.autoSummary.save")}
					</button>
					<!-- 自動要約設定メッセージ -->
					{#if form?.error && isMessageForAction("updateAutoSummarySettings")}
						<div class="mt-3 bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded auto-phrase-target">
							{$_(`settings.messages.${form.error}`) || form.error}
						</div>
					{/if}
					{#if form?.success && isMessageForAction("updateAutoSummarySettings")}
						<div class="mt-3 bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded auto-phrase-target">
							{$_(`settings.messages.${form.message}`) || form.message}
						</div>
					{/if}
							</form>
						</section>
					{/if}

					<!-- LLM処理状況セクション -->
					{#if existingLLMKey}
						<section id="llm-status" class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
							<h3 class="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
								{$_("settings.llmStatus.title")}
							</h3>
				<p class="text-gray-700 dark:text-gray-300 mb-4 auto-phrase-target">
					{$_("settings.llmStatus.description")}
				</p>
				<a
					href="/llm"
					class="inline-flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-md transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
				>
					<svg
						class="w-5 h-5 mr-2"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
						xmlns="http://www.w3.org/2000/svg"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
						></path>
					</svg>
					{$_("settings.llmStatus.viewButton")}
								</a>
						</section>
					{/if}
				</div>
			</div>

			<!-- 危険な操作セクション -->
			<div class="space-y-8">
				<div>
					<h2 class="text-2xl font-bold mb-6 text-red-600 dark:text-red-400">
						{$_("settings.nav.dangerZone")}
					</h2>

					<!-- Account Deletion Section -->
					<section id="delete-account" class="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-red-200 dark:border-red-800">
						<h3 class="text-xl font-semibold mb-4 text-red-600 dark:text-red-400">
							{$_("settings.deleteAccount.title")}
						</h3>
			<div class="mb-4">
				<p class="text-gray-700 dark:text-gray-300 mb-2 auto-phrase-target">
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
			<!-- アカウント削除メッセージ -->
			{#if form?.error && isMessageForAction("deleteAccount")}
				<div class="mt-3 bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded auto-phrase-target">
					{$_(`settings.messages.${form.error}`) || form.error}
				</div>
			{/if}
			{#if form?.success && isMessageForAction("deleteAccount")}
				<div class="mt-3 bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded auto-phrase-target">
					{$_(`settings.messages.${form.message}`) || form.message}
				</div>
			{/if}
				</section>
			</div>
		</div>
	</div>
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
	<p class="text-sm text-gray-500 dark:text-gray-400 auto-phrase-target">
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
	<p class="text-sm text-gray-500 dark:text-gray-400 mb-2 auto-phrase-target">
		{$_("settings.deleteAccount.confirmMessage")}
	</p>
	<p class="text-sm text-red-600 dark:text-red-400 font-medium auto-phrase-target">
		{$_("settings.deleteAccount.warning")}
	</p>
</Modal>