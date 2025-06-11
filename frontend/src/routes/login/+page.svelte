<script lang="ts">
	import { goto } from '$app/navigation';
	import { _ } from 'svelte-i18n';
	import '$lib/i18n';
	import { login } from '$lib/auth';

	let email = '';
	let password = '';
	let error = '';
	let loading = false;

	async function handleLogin() {
		if (!email || !password) {
			error = $_('auth.login.required');
			return;
		}

		loading = true;
		error = '';

		try {
			await login(email, password);
			goto('/');
		} catch (err: any) {
			error = err.message || $_('auth.login.failed');
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-50">
	<div class="max-w-md w-full space-y-8">
		<div>
			<h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
				{$_('auth.login.title')}
			</h2>
		</div>
		<form class="mt-8 space-y-6" on:submit|preventDefault={handleLogin}>
			<div class="space-y-4">
				<div>
					<label for="email" class="sr-only">{$_('auth.login.email')}</label>
					<input
						id="email"
						name="email"
						type="email"
						autocomplete="email"
						required
						class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
						placeholder={$_('auth.login.email')}
						bind:value={email}
					/>
				</div>
				<div>
					<label for="password" class="sr-only">{$_('auth.login.password')}</label>
					<input
						id="password"
						name="password"
						type="password"
						autocomplete="current-password"
						required
						class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
						placeholder={$_('auth.login.password')}
						bind:value={password}
					/>
				</div>
			</div>

			{#if error}
				<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
					{error}
				</div>
			{/if}

			<div>
				<button
					type="submit"
					disabled={loading}
					class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
				>
					{loading ? $_('auth.login.submitting') : $_('auth.login.submit')}
				</button>
			</div>

			<div class="text-center">
				<a href="/register" class="text-indigo-600 hover:text-indigo-500">
					{$_('auth.login.noAccount')}
				</a>
			</div>
		</form>
	</div>
</div>