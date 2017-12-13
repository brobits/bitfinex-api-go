Feature: As an anonymous user, I can request a ticker update
    Scenario: Request a ticker update

        Given I preload the URL "https://api.bitfinex.com/v2/tickerstickers?symbols=tBTCUSD" with the following JSON:
        """
            [
                "tBTCUSD",
                16832,
                48.43302881,
                16833,
                39.28347075,
                -295,
                -0.0172,
                16832,
                51321.37420961,
                17579,
                15800
            ]
        """
        When I visit the URL "https://api.bitfinex.com/v2/tickerstickers?symbols=tBTCUSD"
        Then I should see the following JSON:
        """
            [
                "tBTCUSD",
                16832,
                48.43302881,
                16833,
                39.28347075,
                -295,
                -0.0172,
                16832,
                51321.37420961,
                17579,
                15800
            ]
        """