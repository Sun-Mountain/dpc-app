package gov.cms.dpc.api.client;

import ca.uhn.fhir.context.FhirContext;
import ca.uhn.fhir.parser.IParser;
import ca.uhn.fhir.rest.api.MethodOutcome;
import ca.uhn.fhir.rest.client.api.IClientInterceptor;
import ca.uhn.fhir.rest.client.api.IGenericClient;
import ca.uhn.fhir.rest.client.api.IHttpRequest;
import ca.uhn.fhir.rest.client.api.IHttpResponse;
import ca.uhn.fhir.rest.gclient.IOperationUntypedWithInput;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import gov.cms.dpc.api.models.JobCompletionModel;
import gov.cms.dpc.common.utils.SeedProcessor;
import gov.cms.dpc.fhir.DPCIdentifierSystem;
import gov.cms.dpc.fhir.FHIRExtractors;
import org.apache.commons.lang3.tuple.Pair;
import org.apache.http.HttpHeaders;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.eclipse.jetty.http.HttpStatus;
import org.hl7.fhir.dstu3.model.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.*;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.stream.Collectors;

import static gov.cms.dpc.fhir.FHIRHeaders.PREFER_HEADER;
import static gov.cms.dpc.fhir.FHIRHeaders.PREFER_RESPOND_ASYNC;

/**
 * Shared methods for testing export jobs
 */
public class ClientUtils {
    private static final Logger logger = LoggerFactory.getLogger(ClientUtils.class);

    private ClientUtils() {
        // Not used
    }

    /**
     * Helper method to create a FHIR client has the headers setup for export operations.
     *
     * @param context       - FHIR context to use
     * @param serverBaseURL - the base URL for the FHIR endpoint
     * @param accessToken   - {@link String} access token to use
     * @return {@link IGenericClient} for FHIR requests
     * @see #createExportOperation(IGenericClient, String)
     */
    public static IGenericClient createExportClient(FhirContext context, String serverBaseURL, String accessToken) {
        final IGenericClient exportClient = context.newRestfulGenericClient(serverBaseURL);
        // Add a header the hard way
        final var addPreferInterceptor = new IClientInterceptor() {
            @Override
            public void interceptRequest(IHttpRequest iHttpRequest) {
                iHttpRequest.addHeader(PREFER_HEADER, PREFER_RESPOND_ASYNC);
                if (accessToken != null) {
                    iHttpRequest.addHeader(HttpHeaders.AUTHORIZATION, String.format("Bearer %s", accessToken));
                }
            }

            @Override
            public void interceptResponse(IHttpResponse iHttpResponse) {
                // Not used
            }
        };
        exportClient.registerInterceptor(addPreferInterceptor);
        return exportClient;
    }

    /**
     * Helper method for creating an export request using the FHIR client.
     *
     * @param exportClient - {@link IGenericClient} client to use for the request.
     * @param rosterID     - {@link String} Roster ID to request data for
     * @return - {@link IOperationUntypedWithInput} export request, ready to execute
     */
    private static IOperationUntypedWithInput<Parameters> createExportOperation(IGenericClient exportClient, String rosterID) {
        return exportClient
                .operation()
                .onInstance(new IdType(rosterID))
                .named("$export")
                .withNoParameters(Parameters.class)
                .encodedJson()
                .useHttpGet();
    }

    /**
     * Helper method for creating a roster {@link Bundle} and corresponding FHIR Post
     *
     * @param client            - {@link IGenericClient} client to use for request
     * @param resource          - {@link InputStream} representing test associations file
     * @param organizationID    - {@link UUID} of Organization to submit under
     * @param patientReferences - {@link Map} of patient's associated to a given provider
     * @throws IOException - throws if unable to read the file
     */
    private static void createRosterSubmission(IGenericClient client, InputStream resource, UUID organizationID, Map<String, Reference> patientReferences) throws IOException {

        final Map<String, List<Pair<String, String>>> providerMap = SeedProcessor.extractProviderMap(resource);

        // Find the entry for the given key (yes, I know this is bad)
        providerMap
                .entrySet()
                .forEach(providerRoster -> {

                    final Group attributionRoster = SeedProcessor.generateAttributionGroup(providerRoster, organizationID, patientReferences);

                    // Now, submit the bundle
                    client
                            .create()
                            .resource(attributionRoster)
                            .encodedJson()
                            .execute();
                });
    }

    /**
     * Helper method to wait for an export request to complete, simply polls the given endpoint every second.
     *
     * @param jobLocation   - {@link String} URL where client can get job status
     * @param statusMessage - {@link String} status message to print on each iteration
     * @param token         - {@link String} access token to use
     * @return - {@link JobCompletionModel} Completed job response
     * @throws IOException          - throws if the HTTP request fails
     * @throws InterruptedException - throws if the thread is interrupted
     */
    private static JobCompletionModel awaitExportResponse(String jobLocation, String statusMessage, String token) throws IOException, InterruptedException {
        // Use the traditional HTTP Client to check the job status
        JobCompletionModel jobResponse = null;
        try (CloseableHttpClient client = HttpClients.createDefault()) {
            final HttpGet jobGet = new HttpGet(jobLocation);
            jobGet.setHeader(HttpHeaders.AUTHORIZATION, String.format("Bearer %s", token));
            boolean done = false;

            while (!done) {
                Thread.sleep(1000);
                logger.debug(statusMessage);
                try (CloseableHttpResponse response = client.execute(jobGet)) {
                    final int statusCode = response.getStatusLine().getStatusCode();
                    done = statusCode == HttpStatus.OK_200 || statusCode > 300;
                    if (done) {
                        final ObjectMapper mapper = new ObjectMapper().registerModule(new JavaTimeModule());
                        jobResponse = mapper.readValue(response.getEntity().getContent(), JobCompletionModel.class);
                    }
                }
            }
        }
        return jobResponse;
    }

    /**
     * Helper method to download a file from the {@link gov.cms.dpc.api.resources.v1.DataResource}
     * Uses the {@link File#createTempFile(String, String)} method to create the file handle
     *
     * @param fileID - {@link String} full URL of the file to download
     * @param token  - {@link String} access token to use
     * @return - {@link File} file handle where the data is stored
     * @throws IOException - throws if the HTTP request or file writing fails
     */
    private static File fetchExportedFiles(String fileID, String token) throws IOException {
        try (CloseableHttpClient client = HttpClients.createDefault()) {

            final File tempFile = File.createTempFile("dpc", ".ndjson");

            final HttpGet fileGet = new HttpGet(fileID);
            fileGet.addHeader(HttpHeaders.AUTHORIZATION, "Bearer " + token);
            try (CloseableHttpResponse fileResponse = client.execute(fileGet)) {

                try (FileOutputStream outStream = new FileOutputStream(tempFile)) {
                    fileResponse.getEntity().writeTo(outStream);
                }

                return tempFile;
            }
        }
    }

    /**
     * Helper method for polling the Job Status endpoint until completion.
     * When complete, returns the {@link JobCompletionModel}
     *
     * @param exportOperation - {@link IOperationUntypedWithInput} to execute
     * @param token           - {@link String} access token
     * @return - {@link JobCompletionModel}
     * @throws IOException          - throws if something bad happens
     * @throws InterruptedException - throws if someone cuts in line
     */
    private static JobCompletionModel monitorExportRequest(IOperationUntypedWithInput<Parameters> exportOperation, String token) throws IOException, InterruptedException {
        System.out.println("Retrying export request");

        // Return a MethodOutcome in order to get the response headers.
        final MethodOutcome outcome = exportOperation.returnMethodOutcome().execute();
        // Get the correct header
        final Map<String, List<String>> headers = outcome.getResponseHeaders();

        // Get the headers and check the status
        final String exportURL = headers.get("content-location").get(0);
        System.out.printf("Export job started. Progress URL: %s%n", exportURL);


        // Poll the job until it's done
        return awaitExportResponse(exportURL, "Checking job status", token);
    }

    public static void handleExportJob(IGenericClient exportClient, List<String> providerNPIs, String token) {
        providerNPIs
                .stream()
                .map(npi -> exportClient
                        .search()
                        .forResource(Group.class)
                        .where(Group.CHARACTERISTIC_VALUE
                                .withLeft(Group.CHARACTERISTIC.exactly().systemAndCode("", "attributed-to"))
                                .withRight(Group.VALUE.exactly().systemAndCode(DPCIdentifierSystem.NPPES.getSystem(), npi)))
                        .returnBundle(Bundle.class)
                        .encodedJson()
                        .execute())
                .map(search -> (Group) search.getEntryFirstRep().getResource())
                .map(group -> {
                    final IOperationUntypedWithInput<Parameters> exportOperation = createExportOperation(exportClient, group.getId());
                    try {
                        return monitorExportRequest(exportOperation, token);
                    } catch (IOException | InterruptedException e) {
                        throw new RuntimeException("Error monitoring export", e);
                    }
                })
                .forEach(jobResponse -> jobResponse.getOutput().forEach(entry -> {
                    System.out.println(entry.getUrl());
                    try {
                        final File file = fetchExportedFiles(entry.getUrl(), token);
                        System.out.println(String.format("Downloaded file to: %s", file.getPath()));
                    } catch (IOException e) {
                        throw new RuntimeException("Cannot output file", e);
                    }
                }));
    }

    private static <T extends BaseResource> Bundle bundleSubmitter(Class<?> baseClass, Class<T> clazz, String filename, IParser parser, IGenericClient client) throws IOException {

        try (InputStream resource = baseClass.getClassLoader().getResourceAsStream(filename)) {
            final Bundle bundle = parser.parseResource(Bundle.class, resource);

            final Parameters parameters = new Parameters();
            parameters.addParameter().setResource(bundle);

            return client
                    .operation()
                    .onType(clazz)
                    .named("submit")
                    .withParameters(parameters)
                    .returnResourceType(Bundle.class)
                    .encodedJson()
                    .execute();
        }
    }

    public static Map<String, Reference> submitPatients(Class<?> baseClass, FhirContext ctx, IGenericClient exportClient) {
        final Bundle patientBundle;

        try {
            System.out.println("Submitting patients");
            patientBundle = bundleSubmitter(baseClass, Patient.class, "patient_bundle.json", ctx.newJsonParser(), exportClient);
        } catch (Exception e) {
            throw new RuntimeException("Cannot submit patients.", e);
        }

        final Map<String, Reference> patientReferences = new HashMap<>();
        patientBundle
                .getEntry()
                .stream()
                .map(Bundle.BundleEntryComponent::getResource)
                .map(resource -> (Patient) resource)
                .forEach(patient -> patientReferences.put(patient.getIdentifierFirstRep().getValue(), new Reference(patient.getId())));

        return patientReferences;
    }

    public static List<String> submitPractitioners(Class<?> baseClass, FhirContext ctx, IGenericClient exportClient) {
        final Bundle providerBundle;

        try {
            System.out.println("Submitting practitioners");
            providerBundle = bundleSubmitter(baseClass, Practitioner.class, "provider_bundle.json", ctx.newJsonParser(), exportClient);
        } catch (Exception e) {
            throw new RuntimeException("Cannot submit providers.", e);
        }

        // Get the provider NPIs
        return providerBundle
                .getEntry()
                .stream()
                .map(Bundle.BundleEntryComponent::getResource)
                .map(resource -> (Practitioner) resource)
                .map(FHIRExtractors::getProviderNPI)
                .collect(Collectors.toList());
    }

    public static void createAndUploadRosters(String seedsFile, IGenericClient client, UUID organizationID, Map<String, Reference> patientReferences) throws IOException {
        // Read the provider bundle from the given file
        try (InputStream resource = new FileInputStream(new File(seedsFile))) {
            // Now, submit the bundle
            System.out.println("Uploading Patient roster");
            createRosterSubmission(client, resource, organizationID, patientReferences);
        }
    }
}
