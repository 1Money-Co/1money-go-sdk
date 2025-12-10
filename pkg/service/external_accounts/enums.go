/*
 * Copyright 2025 1Money Co.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package external_accounts

//go:generate go tool go-enum -f=$GOFILE --marshal --names --nocase

// BankNetworkName represents the bank network type for external accounts.
// ENUM(US_ACH, SWIFT, US_FEDWIRE)
type BankNetworkName string

// Currency represents the supported currencies for external accounts.
// ENUM(USD)
type Currency string

// BankAccountStatus represents the status of an external bank account.
// ENUM(PENDING, APPROVED, FAILED)
type BankAccountStatus string

// CountryCode represents ISO 3166-1 alpha-3 country codes.
// ENUM(AND, ARE, AFG, ATG, AIA, ALB, ARM, AGO, ARG, ASM, AUT, AUS, ABW, AZE, BIH, BRB, BGD, BEL, BFA, BGR, BHR, BDI, BEN, BLM, BMU, BRN, BOL, BRA, BHS, BTN, BWA, BLR, BLZ, CAN, CCK, COD, CAF, COG, CHE, CIV, COK, CHL, CMR, CHN, COL, CRI, CUB, CPV, CUW, CXR, CYP, CZE, DEU, DJI, DNK, DMA, DOM, DZA, ECU, EST, EGY, ESH, ERI, ESP, ETH, FIN, FJI, FLK, FSM, FRO, FRA, GAB, GBR, GRD, GEO, GUF, GGY, GHA, GIB, GRL, GMB, GIN, GLP, GNQ, GRC, SGS, GTM, GUM, GNB, GUY, HKG, HND, HRV, HTI, HUN, IDN, IRL, ISR, IMN, IND, IOT, IRQ, IRN, ISL, ITA, JEY, JAM, JOR, JPN, KEN, KGZ, KHM, KIR, COM, KNA, PRK, KOR, KWT, CYM, KAZ, LAO, LBN, LCA, LIE, LKA, LBR, LSO, LTU, LUX, LVA, LBY, MAR, MCO, MDA, MNE, MAF, MDG, MHL, MKD, MLI, MMR, MNG, MAC, MNP, MTQ, MRT, MSR, MLT, MUS, MDV, MWI, MEX, MYS, MOZ, NAM, NCL, NER, NFK, NGA, NIC, NLD, NOR, NPL, NRU, NIU, NZL, OMN, PAN, PER, PYF, PNG, PHL, PAK, POL, SPM, PCN, PRI, PSE, PRT, PLW, PRY, QAT, REU, ROU, SRB, RUS, RWA, SAU, SLB, SYC, SDN, SWE, SGP, SHN, SVN, SJM, SVK, SLE, SMR, SEN, SOM, SUR, SSD, STP, SLV, SXM, SYR, SWZ, TCA, TCD, TGO, THA, TJK, TKL, TLS, TKM, TUN, TON, TUR, TTO, TUV, TWN, TZA, UKR, UGA, USA, URY, UZB, VAT, VCT, VEN, VGB, VIR, VNM, VUT, WLF, WSM, YEM, MYT, ZAF, ZMB, ZWE)
//
// go-enum requires all values on a single line
type CountryCode string
