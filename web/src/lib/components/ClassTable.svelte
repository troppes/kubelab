<script ssr="false">
	import { onMount, onDestroy } from 'svelte';
	import {
		scaleDeployment,
		getConnectionString,
		getStudentsForClass,
		getClasses
	} from '$lib/kubelab-requests.js';
	import DeploymentTable from './DeploymentTable.svelte';

	export let token;
	let classes = [];
	let students = { items: [] };
	let intervalStudent = null;
	let intervalClasses = null;
	let currentClass = null;

	const studentHandler = async (e) => {
		try {
			currentClass = e.srcElement.dataset.id;
			students = await getStudentsForClass(token, currentClass);
			intervalStudent = setInterval(renewStudents, 5000);
			document.querySelector('#details').classList.remove('hidden');
		} catch (error) {
			console.log(error);
		}
	};

	const renewDeployments = async () => {
		classes = await getClasses(token);
	};
	const renewStudents = async () => {
		students = await getStudentsForClass(token, currentClass);
	};

	onDestroy(() => {
		// Clean up the interval when the component is destroyed
		clearInterval(intervalClasses);
	});

	// write onmount to fetch deployments
	onMount(async () => {
		try {
			renewDeployments();
			intervalClasses = setInterval(renewDeployments, 5000);
			classes = await getClasses(token);
		} catch (error) {
			console.log(error);
		}
	});
</script>

<div class="item">
	{#await classes}
		<div>
			<p>Fetching Classrooms ...</p>
		</div>
	{:then classes}
		<div>
			<div>
				<table>
					<thead>
						<tr>
							<th>Name</th>
							<th>File-Manager</th>
							<th>Enrolled Students</th>
							<th>Exam-Mode</th>
						</tr>
					</thead>
					<tbody>
						{#each classes as customClass}
							<tr>
								<td>
									{customClass.metadata.name}
								</td>
								<td>
									<button
										class="button"
										data-id={customClass.metadata.name}
										on:click={console.log('Todo')}>See Files</button
									>
								</td>
								<td>
									<button
										class="button"
										data-id={customClass.metadata.name}
										on:click={studentHandler}>See Students</button
									>
								</td>
								<td>
									{customClass.metadata.annotations.spec.enableExamMode || false}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
		<div id="details" class="hidden">
			<td colspan="4" />
			<DeploymentTable
				{token}
				deployments={students}
				{scaleDeployment}
				{getConnectionString}
				teacherView={true}
			/>
		</div>
	{:catch error}
		<div>
			<p style="color: red">Error loading deployments.</p>
			<p style="color: red">Error message: {error.body.message}</p>
		</div>
	{/await}
</div>

<style>
	.hidden {
		display: none;
	}

	table {
		border-spacing: 10px;
		border-collapse: separate;
		text-align: center;
	}
</style>
