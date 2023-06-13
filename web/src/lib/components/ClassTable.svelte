<script ssr="false">
	import { onMount, onDestroy } from 'svelte';
	import {
		scaleDeployment,
		getConnectionString,
		getStudentsForClass,
		getClasses,
		getFiles,
		deleteFiles,
		uploadFiles
	} from '$lib/kubelab-requests.js';
	import DeploymentTable from './DeploymentTable.svelte';
	import FileManager from './FileManager.svelte';

	export let token;
	let classes = [];
	let students = { items: [] };
	let files = [];
	let intervalStudent = null;
	let intervalClasses = null;
	let intervalFiles = null;
	let currentClass = null;

	const studentHandler = async (e) => {
		try {
			currentClass = e.srcElement.dataset.id;
			students = await getStudentsForClass(token, currentClass);

			clearInterval(intervalFiles);
			intervalStudent = setInterval(renewStudents, 5000);

			document.querySelector('#deployTable').classList.remove('hidden');
			document.querySelector('#fileManager').classList.add('hidden');
		} catch (error) {
			console.log(error);
		}
	};

	const fileHandler = async (e) => {
		try {
			currentClass = e.srcElement.dataset.id;
			files = await getFiles(token, currentClass);

			clearInterval(intervalStudent);
			intervalFiles = setInterval(renewFiles, 5000);

			document.querySelector('#fileManager').classList.remove('hidden');
			document.querySelector('#deployTable').classList.add('hidden');
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
	const renewFiles = async () => {
		files = await getFiles(token, currentClass);
	};

	onDestroy(() => {
		// Clean up the interval when the component is destroyed
		clearInterval(intervalClasses);
		clearInterval(intervalStudent);
		clearInterval(intervalFiles);
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
									<button class="button" data-id={customClass.metadata.name} on:click={fileHandler}
										>See Files</button
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
		<div id="deployTable" class="hidden">
			<DeploymentTable
				{token}
				deployments={students}
				{scaleDeployment}
				{getConnectionString}
				teacherView={true}
			/>
		</div>
		<div id="fileManager" class="hidden">
			<FileManager {token} {files} {currentClass} {uploadFiles} {deleteFiles} />
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
