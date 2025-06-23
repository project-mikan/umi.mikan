<script lang="ts">
import "$lib/i18n";
import type { ActionData } from "./$types.ts";

export let form: ActionData;

let loading = false;
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-50">
	<div class="max-w-md w-full space-y-8">
		<div>
			<h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
				{$_('auth.register.title')}
			</h2>
		</div>
		<form 
			class="mt-8 space-y-6" 
			method="POST" 
			use:enhance={(() => {
				loading = true;
				return async ({ update }) => {
					loading = false;
					await update();
				};
			}) satisfies SubmitFunction}
		>
			{#if form?.error}
				<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
					{form.error}
				</div>
			{/if}
			<div class="space-y-4">
				<div>
					<label for="name" class="sr-only">{$_('auth.register.name')}</label>
					<input
						id="name"
						name="name"
						type="text"
						autocomplete="name"
						required
						class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
						placeholder={$_('auth.register.name')}
					/>
				</div>
				<div>
					<label for="email" class="sr-only">{$_('auth.register.email')}</label>
					<input
						id="email"
						name="email"
						type="email"
						autocomplete="email"
						required
						class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
						placeholder={$_('auth.register.email')}
					/>
				</div>
				<div>
					<label for="password" class="sr-only">{$_('auth.register.password')}</label>
					<input
						id="password"
						name="password"
						type="password"
						autocomplete="new-password"
						required
						class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
						placeholder={$_('auth.register.password')}
					/>
				</div>
			</div>

			<div>
				<button
					type="submit"
					disabled={loading}
					class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
				>
					{loading ? $_('auth.register.submitting') : $_('auth.register.submit')}
				</button>
			</div>

			<div class="text-center">
				<a href="/login" class="text-indigo-600 hover:text-indigo-500">
					{$_('auth.register.hasAccount')}
				</a>
			</div>
		</form>
	</div>
</div>