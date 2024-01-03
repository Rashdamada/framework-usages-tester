package com.square.examples;

import java.io.IOException;
import java.io.InputStream;
import java.util.Properties;

import com.squareup.square.*;
import com.squareup.square.api.*;
import com.squareup.square.exceptions.*;
import com.squareup.square.models.*;
import com.squareup.square.models.Error;

public class App {
    public static void main(String[] args) {

        InputStream inputStream = App.class.getResourceAsStream("/config.properties");
        Properties prop = new Properties();

        try {
            prop.load(inputStream);
        } catch (IOException e) {
            System.out.println("Error reading properties file");
            e.printStackTrace();
        }

        SquareClient client = new SquareClient.Builder()
                .environment(Environment.SANDBOX)
                .accessToken(prop.getProperty("SQUARE_ACCESS_TOKEN"))
                .build();

        LocationsApi locationsApi = client.getLocationsApi();

        locationsApi.listLocationsAsync()
                .thenAccept(result -> {
                    for (Location l : result.getLocations()) {
                        System.out.printf(
                                "%s: %s, %s, %s\n",
                                l.getId(),
                                l.getName(),
                                l.getAddress().getAddressLine1(),
                                l.getAddress().getLocality());
                    }
                })
                .exceptionally(exception -> {
                    try {
                        throw exception.getCause();
                    } catch (ApiException ae) {
                        for (Error err : ae.getErrors()) {
                            System.out.println(err.getCategory());
                            System.out.println(err.getCode());
                            System.out.println(err.getDetail());
                        }
                    } catch (Throwable t) {
                        t.printStackTrace();
                    }
                    return null;
                })
                .join();
        SquareClient.shutdown();
    }
}
