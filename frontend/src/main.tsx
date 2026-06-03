import React, { useEffect, useMemo, useState } from 'react';
import { createRoot } from 'react-dom/client';
import {
  Boxes,
  ClipboardList,
  Pencil,
  Hammer,
  LayoutDashboard,
  PackagePlus,
  ScrollText,
  Shield,
  ShoppingCart,
  Save,
  Trash2,
  Users,
  X
} from 'lucide-react';
import './styles.css';

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8081';

type User = { id: number; playerName: string; role: 'Member' | 'Officer' | 'Admin' | 'Owner'; apiToken?: string };
type BuildItem = { id?: number; slot: string; itemName: string; tier: number; enchantment: number; quantity: number };
type Build = { id: number; name: string; role: string; silverValue: number; screenshotUrl: string; items: BuildItem[] };
type Regear = {
  id: number;
  playerName: string;
  requestDate: string;
  buildId: number;
  buildName: string;
  deathScreenshotUrl: string;
  vodUrl: string;
  notes: string;
  status: string;
  silverValue: number;
  pickupLocation: string;
  items?: RegearItem[];
  createdAt: string;
};
type RegearItem = {
  itemName: string;
  tier: number;
  enchantment: number;
  quantityNeeded: number;
  quantityFulfilled: number;
  quantityMissing: number;
};
type InventoryItem = {
  id: number;
  itemName: string;
  equivalentTier: number;
  quantityAvailable: number;
  lowStockThreshold: number;
  lastUpdated: string;
};
type ShoppingList = { id: number; name: string; status: string; createdAt: string; items: ShoppingListItem[] };
type ShoppingListItem = { itemName: string; equivalentTier: number; quantityNeeded: number };
type DashboardData = {
  pendingRegears: number;
  approvedRegears: number;
  deniedRegears: number;
  pendingSilverValue: number;
  mostRequestedItems: ShoppingListItem[];
  lowStockItems: InventoryItem[];
  recentRegears: Regear[];
  totalInventoryItems: number;
  openShortageQuantity: number;
};
type MemberHistory = { id: number; role: string; playerName: string; requested: number; approved: number; silverValue: number; lastRequestStatus: string };

const slots = ['Main Hand', 'Off Hand', 'Helmet', 'Armor', 'Shoes', 'Cape', 'Food', 'Potion'];
const itemPowers = Array.from({ length: 5 }, (_, tierIndex) =>
  Array.from({ length: 5 }, (_, enchantment) => ({ tier: tierIndex + 4, enchantment }))
).flat();
const buildPowers = [
  { label: 'T7 Equivalent', tier: 7, enchantment: 0 },
  { label: 'T8 Equivalent', tier: 8, enchantment: 0 },
  { label: 'T9 Equivalent', tier: 9, enchantment: 0 }
];
const consumablePowers = [
  { label: '.0', tier: 7, enchantment: 0 },
  { label: '.1', tier: 7, enchantment: 1 },
  { label: '.2', tier: 7, enchantment: 2 },
  { label: '.3', tier: 7, enchantment: 3 }
];
const nav = [
  ['Dashboard', LayoutDashboard],
  ['Regears', ClipboardList],
  ['Builds', Shield],
  ['Inventory', Boxes],
  ['Crafting', Hammer],
  ['Shopping Lists', ShoppingCart],
  ['Members', Users],
  ['Admin Settings', ScrollText]
] as const;

function App() {
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState('Dashboard');
  const [builds, setBuilds] = useState<Build[]>([]);
  const [regears, setRegears] = useState<Regear[]>([]);
  const [inventory, setInventory] = useState<InventoryItem[]>([]);
  const [dashboard, setDashboard] = useState<DashboardData | null>(null);
  const [shoppingList, setShoppingList] = useState<ShoppingList | null>(null);
  const [members, setMembers] = useState<MemberHistory[]>([]);
  const [error, setError] = useState('');

  const client = useMemo(() => makeClient(setError), []);
  const canManage = ['Officer', 'Admin', 'Owner'].includes(user?.role as string);

  useEffect(() => {
    // Check if user is already authenticated via stored token
    const token = localStorage.getItem('ao_token');
    if (!token) {
      setIsLoading(false);
      return;
    }
    client.setToken(token);
    client.get<User>('/api/auth/me')
      .then((userData) => {
        setUser(userData);
        setIsAuthenticated(true);
      })
      .catch(() => {
        // Token is invalid, clear it
        localStorage.removeItem('ao_token');
        client.setToken('');
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, [client]);

  useEffect(() => {
    if (!user || !isAuthenticated) return;
    refresh();
  }, [user, isAuthenticated]);

  async function handleLogin(username: string, password: string) {
    try {
      setError('');
      const userData = await client.post<User>('/api/auth/login', { playerName: username, password });
      if (userData.apiToken) {
        localStorage.setItem('ao_token', userData.apiToken);
        client.setToken(userData.apiToken);
      }
      setUser(userData);
      setIsAuthenticated(true);
    } catch (err) {
      // Error is already set by the client
      throw err;
    }
  }

  async function handleSignup(username: string, password: string) {
    try {
      setError('');
      const userData = await client.post<User>('/api/auth/signup', { playerName: username, password });
      if (userData.apiToken) {
        localStorage.setItem('ao_token', userData.apiToken);
        client.setToken(userData.apiToken);
      }
      setUser(userData);
      setIsAuthenticated(true);
    } catch (err) {
      // Error is already set by the client
      throw err;
    }
  }

  async function handleLogout() {
    try {
      await client.post('/api/auth/logout', {});
    } catch (err) {
      // Ignore logout errors
    }
    localStorage.removeItem('ao_token');
    client.setToken('');
    setUser(null);
    setIsAuthenticated(false);
    setError('');
  }

  async function refresh() {
    const [buildRows, regearRows, inventoryRows, dash] = await Promise.all([
      client.get<Build[]>('/api/builds'),
      client.get<Regear[]>('/api/regears'),
      client.get<InventoryItem[]>('/api/inventory'),
      client.get<DashboardData>('/api/dashboard')
    ]);
    setBuilds(buildRows);
    setRegears(regearRows);
    setInventory(inventoryRows);
    setDashboard(dash);
    if (canManage) client.get<MemberHistory[]>('/api/members/history').then(setMembers).catch(() => undefined);
    client.getOptional<ShoppingList>('/api/shopping-lists/latest').then(setShoppingList);
  }

  if (isLoading) return <Loading error="" />;
  if (!isAuthenticated || !user) return <AuthScreen onLogin={handleLogin} onSignup={handleSignup} error={error} />;

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="brand">
          <div className="brandMark">AO</div>
          <div>
            <strong>AO Guild Management</strong>
            <span>{user.role} console</span>
          </div>
        </div>
        <nav>
          {nav
            .filter(([label]) => canManage || !['Inventory', 'Crafting', 'Shopping Lists', 'Members', 'Admin Settings'].includes(label as string))
            .map(([label, Icon]) => (
            <button key={label} className={page === label ? 'active' : ''} onClick={() => setPage(label)}>
              <Icon size={18} />
              <span>{label}</span>
            </button>
          ))}
        </nav>
        <div className="userPill">
          <Shield size={18} />
          <span>{user.playerName}</span>
          <button className="logoutBtn" onClick={handleLogout} title="Logout">
            <X size={16} />
          </button>
        </div>
      </aside>

      <main className="main">
        <header className="topbar">
          <div>
            <p>Albion Online Guild Regear Management</p>
            <h1>{page}</h1>
          </div>
          <button className="primary" onClick={refresh}>
            Refresh
          </button>
        </header>
        {error && <div className="alert">{error}</div>}
        {page === 'Dashboard' && dashboard && <Dashboard data={dashboard} canManage={canManage} />}
        {page === 'Regears' && (
          <Regears builds={builds} regears={regears} canManage={canManage} client={client} refresh={refresh} user={user} />
        )}
        {page === 'Builds' && <Builds builds={builds} canManage={canManage} client={client} refresh={refresh} />}
        {page === 'Inventory' && canManage && <Inventory inventory={inventory} canManage={canManage} client={client} refresh={refresh} />}
        {page === 'Crafting' && canManage && <Crafting shoppingList={shoppingList} />}
        {page === 'Shopping Lists' && canManage && <ShoppingLists list={shoppingList} canManage={canManage} client={client} refresh={refresh} />}
        {page === 'Members' && <Members members={members} currentUser={user!} canManage={canManage} client={client} refresh={refresh} />}
        {page === 'Admin Settings' && canManage && <AdminSettings user={user} />}
      </main>
    </div>
  );
}

function AuthScreen({ onLogin, onSignup, error }: { 
  onLogin: (username: string, password: string) => Promise<void>; 
  onSignup: (username: string, password: string) => Promise<void>; 
  error: string 
}) {
  const [isSignup, setIsSignup] = useState(false);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    if (!username.trim() || !password.trim()) return;
    
    setIsLoading(true);
    try {
      if (isSignup) {
        await onSignup(username.trim(), password);
      } else {
        await onLogin(username.trim(), password);
      }
    } catch (err) {
      // Error is handled by the parent component
    } finally {
      setIsLoading(false);
    }
  }

  function switchMode() {
    setIsSignup(!isSignup);
    setUsername('');
    setPassword('');
  }

  return (
    <main className="login">
      <section className="loginPanel">
        <div className="brandMark large">AO</div>
        <h1>Guild Regear Hall</h1>
        <p>{isSignup ? 'Create your account to join the guild' : 'Sign in to access the guild management system'}</p>
        
        {error && <div className="alert">{error}</div>}
        
        {!isSignup && (
          <div className="demoCredentials">
            <h3>Admin Access</h3>
            <p><strong>Username:</strong> Blazor</p>
            <p><strong>Password:</strong> 123</p>
          </div>
        )}
        
        <form onSubmit={handleSubmit}>
          <div className="field">
            <label>{isSignup ? 'In-Game Name' : 'Username'}</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder={isSignup ? 'Enter your Albion Online character name' : 'Enter your username'}
              disabled={isLoading}
              autoComplete="username"
              required
            />
          </div>
          
          <div className="field">
            <label>Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder={isSignup ? 'Create a password (min 3 characters)' : 'Enter your password'}
              disabled={isLoading}
              autoComplete={isSignup ? 'new-password' : 'current-password'}
              required
              minLength={isSignup ? 3 : undefined}
            />
          </div>
          
          <button 
            type="submit" 
            className="primary loginBtn"
            disabled={isLoading || !username.trim() || !password.trim() || (isSignup && password.length < 3)}
          >
            {isLoading ? (isSignup ? 'Creating Account...' : 'Signing In...') : (isSignup ? 'Create Account' : 'Sign In')}
          </button>
        </form>

        <div className="authSwitch">
          {isSignup ? (
            <p>
              Already have an account? 
              <button type="button" className="linkBtn" onClick={switchMode}>
                Sign In
              </button>
            </p>
          ) : (
            <p>
              Don't have an account? 
              <button type="button" className="linkBtn" onClick={switchMode}>
                Sign Up
              </button>
            </p>
          )}
        </div>
      </section>
    </main>
  );
}

function LoginScreen({ onLogin, error }: { onLogin: (username: string, password: string) => Promise<void>; error: string }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    if (!username.trim() || !password.trim()) return;
    
    setIsLoading(true);
    try {
      await onLogin(username.trim(), password);
    } catch (err) {
      // Error is handled by the parent component
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <main className="login">
      <section className="loginPanel">
        <div className="brandMark large">AO</div>
        <h1>Guild Regear Hall</h1>
        <p>Sign in to access the guild management system</p>
        
        {error && <div className="alert">{error}</div>}
        
        <div className="demoCredentials">
          <h3>Demo Credentials</h3>
          <p><strong>Username:</strong> Blazor</p>
          <p><strong>Password:</strong> 123</p>
        </div>
        
        <form onSubmit={handleSubmit}>
          <div className="field">
            <label>Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Enter your username"
              disabled={isLoading}
              autoComplete="username"
              required
            />
          </div>
          
          <div className="field">
            <label>Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              disabled={isLoading}
              autoComplete="current-password"
              required
            />
          </div>
          
          <button 
            type="submit" 
            className="primary loginBtn"
            disabled={isLoading || !username.trim() || !password.trim()}
          >
            {isLoading ? 'Signing In...' : 'Sign In'}
          </button>
        </form>
      </section>
    </main>
  );
}

function Loading({ error }: { error: string }) {
  return (
    <main className="login">
      <section className="loginPanel">
        <div className="brandMark large">AO</div>
        <h1>Guild Regear Hall</h1>
        <p>{error || 'Loading...'}</p>
      </section>
    </main>
  );
}

function Dashboard({ data, canManage }: { data: DashboardData; canManage: boolean }) {
  return (
    <>
      <section className="statsGrid">
        <Stat label="Pending" value={data.pendingRegears} />
        <Stat label="Approved" value={data.approvedRegears} />
        <Stat label="Denied" value={data.deniedRegears} />
        {canManage && <Stat label="Open Shortages" value={data.openShortageQuantity} />}
      </section>
      {canManage && (
        <Panel title="Low Stock">
          <Table
            columns={['Item', 'Qty', 'Floor']}
            rows={(data.lowStockItems ?? []).map((i) => [equivalentItemLabel(i), i.quantityAvailable, i.lowStockThreshold])}
          />
        </Panel>
      )}
      <Panel title="Recent Regears">
        <RegearTable rows={data.recentRegears ?? []} />
      </Panel>
    </>
  );
}

function parsePickupLocation(loc: string): { column: string; row: string } {
  if (!loc) return { column: '', row: '' };
  const parts = loc.split(',');
  if (parts.length === 2) {
    return { column: parts[0], row: parts[1] };
  }
  return { column: loc, row: '' };
}

function serializePickupLocation(column: string, row: string): string {
  return `${column.trim()},${row.trim()}`;
}

function isPickupLocationSet(loc: string): boolean {
  if (!loc) return false;
  const parts = loc.split(',');
  if (parts.length === 2) {
    return parts[0].trim() !== '' && parts[1].trim() !== '';
  }
  return false;
}

function Regears({ builds, regears, canManage, client, refresh, user }: {
  builds: Build[];
  regears: Regear[];
  canManage: boolean;
  client: ReturnType<typeof makeClient>;
  refresh: () => Promise<void>;
  user: User;
}) {
  const [form, setForm] = useState({ playerName: user.playerName, buildId: builds[0]?.id ?? 1, deathScreenshotUrl: '', vodUrl: '', notes: '' });
  const [submitted, setSubmitted] = useState(false);

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    await client.post('/api/regears', { ...form, buildId: Number(form.buildId), requestDate: new Date().toISOString().slice(0, 10) });
    setSubmitted(true);
    setForm({ ...form, deathScreenshotUrl: '', vodUrl: '', notes: '' });
    await refresh();
    // Reset submitted status after 3 seconds
    setTimeout(() => setSubmitted(false), 3000);
  }

  async function status(id: number, value: string, pickupLocation?: string) {
    await client.patch(`/api/regears/${id}/status`, { status: value, pickupLocation: pickupLocation ?? '' });
    await refresh();
  }

  if (canManage) {
    // Admin/Officer view - can review all regear requests
    const pendingRequests = regears.filter((r) => r.status === 'Pending');
    const approvedNeedsLocation = regears.filter((r) => r.status === 'Approved' && !isPickupLocationSet(r.pickupLocation));
    const awaitingPickup = regears.filter((r) => r.status === 'Approved' && isPickupLocationSet(r.pickupLocation));

    return (
      <section className="twoCol wideFirst">
        <Panel title="Regear Requests - Admin Review">
          <div className="queue">
            <h3 className="queueSectionHeader">Admin Review ({pendingRequests.length})</h3>
            {pendingRequests.map((r) => (
              <article key={r.id} className="requestRow">
                <div className="requestGrid">
                  <div>
                    <strong style={{ display: 'block', fontSize: '15px' }}>{r.playerName}</strong>
                    <div style={{ color: 'var(--text-muted)', fontSize: '12px', marginTop: '2px' }}>
                      {r.buildName} &bull; Requested: {new Date(r.requestDate).toLocaleDateString()}
                    </div>
                    {r.notes && <p className="requestNotes">{r.notes}</p>}
                    <RegearPreviews deathScreenshotUrl={r.deathScreenshotUrl} vodUrl={r.vodUrl} />
                  </div>
                  <div className="requestControls">
                    <StatusBadge value={r.status} />
                    <div className="adminActions">
                      <button className="approveBtn" onClick={() => status(r.id, 'Approved')}>Approve</button>
                      <button className="denyBtn" onClick={() => status(r.id, 'Denied')}>Deny</button>
                    </div>
                  </div>
                </div>
              </article>
            ))}
            {pendingRequests.length === 0 && <p className="muted">No requests awaiting review.</p>}

            <h3 className="queueSectionHeader" style={{ marginTop: '28px' }}>Approved - Needs Location ({approvedNeedsLocation.length})</h3>
            {approvedNeedsLocation.map((r) => (
              <article key={r.id} className="requestRow">
                <div className="requestGrid">
                  <div>
                    <strong style={{ display: 'block', fontSize: '15px' }}>{r.playerName}</strong>
                    <div style={{ color: 'var(--text-muted)', fontSize: '12px', marginTop: '2px' }}>
                      {r.buildName} &bull; Requested: {new Date(r.requestDate).toLocaleDateString()}
                    </div>
                    {r.notes && <p className="requestNotes">{r.notes}</p>}
                    <RegearPreviews deathScreenshotUrl={r.deathScreenshotUrl} vodUrl={r.vodUrl} />
                  </div>
                  <div className="requestControls">
                    <StatusBadge value={r.status} />
                    <label className="locationLabel">Set Pickup Location</label>
                    <div className="pickupLocationInputs">
                      <input
                        key={`col-${r.id}-${r.pickupLocation}`}
                        className="locationInput"
                        placeholder="Column"
                        defaultValue={parsePickupLocation(r.pickupLocation).column}
                        onBlur={(e) => {
                          const { row } = parsePickupLocation(r.pickupLocation);
                          const newVal = serializePickupLocation(e.target.value, row);
                          if (newVal !== r.pickupLocation) status(r.id, r.status, newVal);
                        }}
                      />
                      <input
                        key={`row-${r.id}-${r.pickupLocation}`}
                        className="locationInput"
                        placeholder="Row"
                        defaultValue={parsePickupLocation(r.pickupLocation).row}
                        onBlur={(e) => {
                          const { column } = parsePickupLocation(r.pickupLocation);
                          const newVal = serializePickupLocation(column, e.target.value);
                          if (newVal !== r.pickupLocation) status(r.id, r.status, newVal);
                        }}
                      />
                    </div>
                  </div>
                </div>
              </article>
            ))}
            {approvedNeedsLocation.length === 0 && <p className="muted">No approved requests awaiting location.</p>}

            <h3 className="queueSectionHeader" style={{ marginTop: '28px' }}>Completed / Awaiting Pickup ({awaitingPickup.length})</h3>
            {awaitingPickup.map((r) => {
              const { column, row } = parsePickupLocation(r.pickupLocation);
              return (
                <article key={r.id} className="requestRow">
                  <div className="requestGrid">
                    <div>
                      <strong style={{ display: 'block', fontSize: '15px' }}>{r.playerName}</strong>
                      <div style={{ color: 'var(--text-muted)', fontSize: '12px', marginTop: '2px' }}>
                        {r.buildName} &bull; Requested: {new Date(r.requestDate).toLocaleDateString()}
                      </div>
                      {r.notes && <p className="requestNotes">{r.notes}</p>}
                      <RegearPreviews deathScreenshotUrl={r.deathScreenshotUrl} vodUrl={r.vodUrl} />
                    </div>
                    <div className="requestControls">
                      <StatusBadge value={r.status} />
                      <div className="pickupLocationBox">
                        <div className="pickupLocationLabel">📦 Pickup Location</div>
                        <div className="pickupLocationGrid">
                          <div className="pickupLocationCell">
                            <span>Column</span>
                            <strong>{column || '—'}</strong>
                          </div>
                          <div className="pickupLocationCell">
                            <span>Row</span>
                            <strong>{row || '—'}</strong>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </article>
              );
            })}
            {awaitingPickup.length === 0 && <p className="muted">No requests awaiting pickup.</p>}
          </div>
        </Panel>
        <Panel title="Submit New Regear">
          {submitted && <div className="successMessage">Regear request submitted successfully!</div>}
          <form className="formGrid" onSubmit={submit}>
            <input value={form.playerName} onChange={(e) => setForm({ ...form, playerName: e.target.value })} placeholder="Player name" />
            <select value={form.buildId} onChange={(e) => setForm({ ...form, buildId: Number(e.target.value) })}>
              {builds.map((b) => <option key={b.id} value={b.id}>{b.name}</option>)}
            </select>
            <input value={form.deathScreenshotUrl} onChange={(e) => setForm({ ...form, deathScreenshotUrl: e.target.value })} placeholder="Death screenshot URL" />
            <input value={form.vodUrl} onChange={(e) => setForm({ ...form, vodUrl: e.target.value })} placeholder="VOD URL (YouTube, Medal, etc.)" />
            <textarea value={form.notes} onChange={(e) => setForm({ ...form, notes: e.target.value })} placeholder="Notes" />
            <button className="primary">Submit Request</button>
          </form>
        </Panel>
      </section>
    );
  } else {
    // Member view - can only submit requests and see their own
    const userRegears = regears.filter(r => r.playerName.toLowerCase() === user.playerName.toLowerCase());
    
    return (
      <section className="twoCol wideFirst">
        <Panel title="My Regear Requests">
          <div className="queue">
            {userRegears.map((r) => (
              <article key={r.id} className="requestRow memberView">
                <div className="requestGrid">
                  <div>
                    <strong style={{ display: 'block', fontSize: '15px' }}>{r.buildName}</strong>
                    <div style={{ color: 'var(--text-muted)', fontSize: '12px', marginTop: '2px' }}>
                      Requested: {new Date(r.requestDate).toLocaleDateString()}
                    </div>
                    {r.notes && <p className="requestNotes">{r.notes}</p>}
                    <RegearPreviews deathScreenshotUrl={r.deathScreenshotUrl} vodUrl={r.vodUrl} />
                  </div>
                  <div className="requestControls">
                    <StatusBadge value={r.status === 'Approved' && !isPickupLocationSet(r.pickupLocation) ? 'Approved - Waiting for Location' : r.status} />
                    {isPickupLocationSet(r.pickupLocation) && (() => {
                      const { column, row } = parsePickupLocation(r.pickupLocation);
                      return (
                        <div className="pickupLocationBox">
                          <div className="pickupLocationLabel">📦 Pickup Location</div>
                          <div className="pickupLocationGrid">
                            <div className="pickupLocationCell">
                              <span>Column</span>
                              <strong>{column || '—'}</strong>
                            </div>
                            <div className="pickupLocationCell">
                              <span>Row</span>
                              <strong>{row || '—'}</strong>
                            </div>
                          </div>
                        </div>
                      );
                    })()}
                    {r.status === 'Approved' && isPickupLocationSet(r.pickupLocation) && (
                      <button
                        className="claimButton"
                        onClick={() => status(r.id, 'Completed')}
                      >
                        Claim Regear
                      </button>
                    )}
                  </div>
                </div>
              </article>
            ))}
            {userRegears.length === 0 && <p className="muted">You have no regear requests yet.</p>}
          </div>
        </Panel>
        <Panel title="Submit Regear Request">
          {submitted && <div className="successMessage">Regear request submitted successfully!</div>}
          <form className="formGrid" onSubmit={submit}>
            <label className="field">
              <span>Build</span>
              <select value={form.buildId} onChange={(e) => setForm({ ...form, buildId: Number(e.target.value) })}>
                {builds.map((b) => <option key={b.id} value={b.id}>{b.name}</option>)}
              </select>
            </label>
            <label className="field">
              <span>Death Screenshot URL</span>
              <input value={form.deathScreenshotUrl} onChange={(e) => setForm({ ...form, deathScreenshotUrl: e.target.value })} placeholder="Death screenshot URL" />
            </label>
            <label className="field">
              <span>VOD URL</span>
              <input value={form.vodUrl} onChange={(e) => setForm({ ...form, vodUrl: e.target.value })} placeholder="YouTube, Medal, Twitch clip link, etc." />
            </label>
            <label className="field">
              <span>Notes</span>
              <textarea value={form.notes} onChange={(e) => setForm({ ...form, notes: e.target.value })} placeholder="Additional notes about the death" />
            </label>
            <button className="primary">Submit Regear Request</button>
          </form>
        </Panel>
      </section>
    );
  }
}

function Builds({ builds, canManage, client, refresh }: { builds: Build[]; canManage: boolean; client: ReturnType<typeof makeClient>; refresh: () => Promise<void> }) {
  const [editingId, setEditingId] = useState<number | null>(null);
  const [buildToRemove, setBuildToRemove] = useState<Build | null>(null);
  const [viewingBuild, setViewingBuild] = useState<Build | null>(null);
  const [name, setName] = useState('');
  const [role, setRole] = useState('Tank');
  const [silverValue, setSilverValue] = useState(1500000);
  const [screenshotUrl, setScreenshotUrl] = useState('');
  const [items, setItems] = useState<BuildItem[]>(slots.map((slot) => ({ 
    slot, 
    itemName: '', 
    tier: ['Main Hand', 'Off Hand', 'Helmet', 'Armor', 'Shoes'].includes(slot) ? 8 : 7, 
    enchantment: ['Main Hand', 'Off Hand', 'Helmet', 'Armor', 'Shoes'].includes(slot) ? 0 : 1, 
    quantity: 1 
  })));

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    const payload = { name, role, silverValue, screenshotUrl, items: items.filter((i) => i.itemName.trim()) };
    if (editingId) {
      await client.patch(`/api/builds/${editingId}`, payload);
    } else {
      await client.post('/api/builds', payload);
    }
    resetForm();
    await refresh();
  }

  function edit(build: Build) {
    setEditingId(build.id);
    setName(build.name);
    setRole(build.role);
    setSilverValue(build.silverValue);
    setScreenshotUrl(build.screenshotUrl ?? '');
    setItems(slots.map((slot) => {
      const existing = build.items.find((item) => item.slot === slot);
      const defaultTier = ['Main Hand', 'Off Hand', 'Helmet', 'Armor', 'Shoes'].includes(slot) ? 8 : 7;
      const defaultEnchantment = ['Main Hand', 'Off Hand', 'Helmet', 'Armor', 'Shoes'].includes(slot) ? 0 : 1;
      return existing ?? { slot, itemName: '', tier: defaultTier, enchantment: defaultEnchantment, quantity: 1 };
    }));
  }

  async function remove(build: Build) {
    await client.delete(`/api/builds/${build.id}`);
    setBuildToRemove(null);
    if (editingId === build.id) resetForm();
    await refresh();
  }

  function resetForm() {
    setEditingId(null);
    setName('');
    setRole('Tank');
    setSilverValue(1500000);
    setScreenshotUrl('');
    setItems(slots.map((slot) => ({ 
      slot, 
      itemName: '', 
      tier: ['Main Hand', 'Off Hand', 'Helmet', 'Armor', 'Shoes'].includes(slot) ? 8 : 7, 
      enchantment: ['Main Hand', 'Off Hand', 'Helmet', 'Armor', 'Shoes'].includes(slot) ? 0 : 1, 
      quantity: 1 
    })));
  }

  return (
    <section className={canManage ? "twoCol wideFirst" : ""}>
      <Panel title="Approved Builds">
        <div className="buildGrid">
          {builds.map((build) => (
            <article key={build.id} className="buildCard" style={{ cursor: 'pointer' }} onClick={() => setViewingBuild(build)}>
              <div className="buildHeader">
                <div>
                  <strong>{build.name}</strong>
                  <span>{build.role}</span>
                </div>
                {canManage && (
                  <div className="iconActions">
                    <button type="button" title="Edit build" onClick={(e) => { e.stopPropagation(); edit(build); }}><Pencil size={16} /></button>
                    <button type="button" title="Remove build" onClick={(e) => { e.stopPropagation(); setBuildToRemove(build); }}><Trash2 size={16} /></button>
                  </div>
                )}
              </div>
              <div className="buildBody">
                <ul>{build.items.map((i) => <li key={i.slot}>{i.slot}: {buildItemLabel(i)}</li>)}</ul>
                {build.screenshotUrl && <img className="buildScreenshot" src={build.screenshotUrl} alt={`${build.name} build screenshot`} />}
              </div>
            </article>
          ))}
          {builds.length === 0 && <p className="muted">No approved builds yet.</p>}
        </div>
      </Panel>
      {canManage && (
        <Panel title={editingId ? 'Edit Build' : 'Create Build'}>
          <form className="formGrid" onSubmit={submit}>
            <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Build name" />
            <select value={role} onChange={(e) => setRole(e.target.value)}>
              {['Tank', 'Healer', 'DPS', 'Support'].map((r) => <option key={r}>{r}</option>)}
            </select>
            <label className="filePicker">
              <span>Build Screenshot</span>
              <input type="file" accept="image/*" onChange={(e) => loadScreenshot(e, setScreenshotUrl)} />
            </label>
            {screenshotUrl && (
              <div className="screenshotPreview">
                <img src={screenshotUrl} alt="Build screenshot preview" />
                <button className="secondary" type="button" onClick={() => setScreenshotUrl('')}>Remove Screenshot</button>
              </div>
            )}
            {items.map((item, index) => {
              const isConsumable = item.slot === 'Food' || item.slot === 'Potion';
              const powerOptions = isConsumable ? consumablePowers : buildPowers;
              
              return (
                <div className="slotRow" key={item.slot}>
                  <span>{item.slot}</span>
                  <select
                    value={`${item.tier}.${item.enchantment}`}
                    onChange={(e) => {
                      const [tier, enchantment] = e.target.value.split('.').map(Number);
                      setItems(updateAt(items, index, { ...item, tier, enchantment }));
                    }}
                  >
                    {powerOptions.map((power) => (
                      <option key={`${item.slot}-${power.tier}.${power.enchantment}`} value={`${power.tier}.${power.enchantment}`}>
                        {power.label}
                      </option>
                    ))}
                  </select>
                  <input value={item.itemName} onChange={(e) => setItems(updateAt(items, index, { ...item, itemName: e.target.value }))} placeholder="Item" />
                </div>
              );
            })}
            <div className="formActions">
               <button className="primary">{editingId ? 'Save Changes' : 'Save Build'}</button>
               {editingId && <button className="secondary" type="button" onClick={resetForm}>Cancel</button>}
            </div>
          </form>
        </Panel>
      )}
      {buildToRemove && (
        <ConfirmDialog
          title="Remove Build"
          message={`Are you sure you want to remove ${buildToRemove.name}?`}
          confirmLabel="Yes, Remove"
          cancelLabel="No, Keep It"
          onConfirm={() => remove(buildToRemove)}
          onCancel={() => setBuildToRemove(null)}
        />
      )}
      {viewingBuild && <BuildModal build={viewingBuild} onClose={() => setViewingBuild(null)} />}
    </section>
  );
}

function BuildModal({ build, onClose }: { build: Build; onClose: () => void }) {
  return (
    <div className="modalBackdrop" role="presentation" onMouseDown={onClose}>
      <section
        className="buildModalContent"
        role="dialog"
        aria-modal="true"
        onMouseDown={(event) => event.stopPropagation()}
        style={{ maxWidth: '900px', width: '95%', maxHeight: '90vh', display: 'flex', flexDirection: 'column', background: 'var(--panel-bg, #1e1e24)', padding: '24px', borderRadius: '12px' }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '20px' }}>
          <div>
            <h2 style={{ margin: 0, fontSize: '24px' }}>{build.name}</h2>
            <div style={{ color: 'var(--text-muted)' }}>{build.role}</div>
          </div>
          <button type="button" title="Close" style={{ background: 'none', border: 'none', color: 'var(--text)', cursor: 'pointer' }} onClick={onClose}><X size={24} /></button>
        </div>
        
        <div className="buildModalBody">
          {build.screenshotUrl ? (
            <div className="buildModalImage">
              <img 
                src={build.screenshotUrl} 
                alt={`${build.name} build screenshot`} 
              />
            </div>
          ) : (
            <div className="buildModalImage" style={{ background: 'rgba(0,0,0,0.2)', height: '200px', borderRadius: '8px', border: '1px dashed rgba(255,255,255,0.1)', display: 'grid', placeItems: 'center' }}>
              <span className="muted">No build image available</span>
            </div>
          )}
          
          <div className="buildModalItems">
            <h3>Required Items</h3>
            <ul className="buildModalList">
              {build.items.map((i) => (
                <li key={i.slot}>
                  <div>
                    <span>{i.slot}</span>
                    <strong>{buildItemLabel(i)}</strong>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </section>
    </div>
  );
}

function ConfirmDialog({ title, message, confirmLabel, cancelLabel, onConfirm, onCancel }: {
  title: string;
  message: string;
  confirmLabel: string;
  cancelLabel: string;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  return (
    <div className="modalBackdrop" role="presentation" onMouseDown={onCancel}>
      <section
        className="confirmDialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="confirm-title"
        onMouseDown={(event) => event.stopPropagation()}
      >
        <div className="modalIcon"><Trash2 size={20} /></div>
        <h2 id="confirm-title">{title}</h2>
        <p>{message}</p>
        <div className="modalActions">
          <button className="secondary" type="button" onClick={onCancel}>{cancelLabel}</button>
          <button className="danger" type="button" onClick={onConfirm}>{confirmLabel}</button>
        </div>
      </section>
    </div>
  );
}

function Inventory({ inventory, canManage, client, refresh }: { inventory: InventoryItem[]; canManage: boolean; client: ReturnType<typeof makeClient>; refresh: () => Promise<void> }) {
  const [item, setItem] = useState({ itemName: '', equivalentTier: 8, quantityAvailable: 1, lowStockThreshold: 5 });
  const [editingId, setEditingId] = useState<number | null>(null);
  const [stockToRemove, setStockToRemove] = useState<InventoryItem | null>(null);
  const [draft, setDraft] = useState({ quantityAvailable: 0, lowStockThreshold: 0 });

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    await client.post('/api/inventory', item);
    setItem({ ...item, itemName: '' });
    await refresh();
  }

  function edit(stock: InventoryItem) {
    setEditingId(stock.id);
    setDraft({ quantityAvailable: stock.quantityAvailable, lowStockThreshold: stock.lowStockThreshold });
  }

  async function save(stock: InventoryItem) {
    await client.patch(`/api/inventory/${stock.id}`, { ...stock, ...draft });
    setEditingId(null);
    await refresh();
  }

  async function remove(stock: InventoryItem) {
    await client.delete(`/api/inventory/${stock.id}`);
    setStockToRemove(null);
    if (editingId === stock.id) setEditingId(null);
    await refresh();
  }

  return (
    <section className="twoCol wideFirst">
      <Panel title="Guild Chest Stock">
        <div className="tableWrap">
          <table>
            <thead>
              <tr>
                <th>Item</th>
                <th>Equivalent</th>
                <th>Available</th>
                <th>Low</th>
                {canManage && <th>Actions</th>}
              </tr>
            </thead>
            <tbody>
              {inventory.map((stock) => {
                const isEditing = editingId === stock.id;
                return (
                  <tr key={stock.id}>
                    <td>{stock.itemName}</td>
                    <td>T{stock.equivalentTier} Equivalent</td>
                    <td>
                      {isEditing ? (
                        <input
                          className="tableInput"
                          type="number"
                          min="0"
                          value={draft.quantityAvailable}
                          onChange={(e) => setDraft({ ...draft, quantityAvailable: Number(e.target.value) })}
                        />
                      ) : stock.quantityAvailable}
                    </td>
                    <td>
                      {isEditing ? (
                        <input
                          className="tableInput"
                          type="number"
                          min="0"
                          value={draft.lowStockThreshold}
                          onChange={(e) => setDraft({ ...draft, lowStockThreshold: Number(e.target.value) })}
                        />
                      ) : stock.lowStockThreshold}
                    </td>
                    {canManage && (
                      <td>
                        {isEditing ? (
                          <div className="iconActions">
                            <button type="button" title="Save stock" onClick={() => save(stock)}><Save size={16} /></button>
                            <button type="button" title="Cancel" onClick={() => setEditingId(null)}><X size={16} /></button>
                          </div>
                        ) : (
                          <div className="iconActions">
                            <button type="button" title="Edit stock" onClick={() => edit(stock)}><Pencil size={16} /></button>
                            <button type="button" title="Remove stock" onClick={() => setStockToRemove(stock)}><Trash2 size={16} /></button>
                          </div>
                        )}
                      </td>
                    )}
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </Panel>
      <Panel title="Update Stock">
        {canManage ? (
          <form className="formGrid" onSubmit={submit}>
            <label className="field">
              <span>Item Name</span>
              <input value={item.itemName} onChange={(e) => setItem({ ...item, itemName: e.target.value })} placeholder="Item name" />
            </label>
            <label className="field">
              <span>Equivalent Tier</span>
              <select
                value={item.equivalentTier}
                onChange={(e) => setItem({ ...item, equivalentTier: Number(e.target.value) })}
              >
                {Array.from({ length: 7 }, (_, i) => i + 4).map((equiv) => (
                  <option key={`stock-equiv-${equiv}`} value={equiv}>
                    T{equiv} Equivalent
                  </option>
                ))}
              </select>
            </label>
            <label className="field">
              <span>Quantity Available</span>
              <input type="number" value={item.quantityAvailable} onChange={(e) => setItem({ ...item, quantityAvailable: Number(e.target.value) })} />
            </label>
            <label className="field">
              <span>Low Stock Alert</span>
              <input type="number" value={item.lowStockThreshold} onChange={(e) => setItem({ ...item, lowStockThreshold: Number(e.target.value) })} />
            </label>
            <button className="primary"><PackagePlus size={16} /> Save Stock</button>
          </form>
        ) : <p className="muted">Officer permission required.</p>}
      </Panel>
      {stockToRemove && (
        <ConfirmDialog
          title="Remove Stock"
          message={`Are you sure you want to remove T${stockToRemove.equivalentTier} Equivalent ${stockToRemove.itemName} from the guild chest stock?`}
          confirmLabel="Yes, Remove"
          cancelLabel="No, Keep It"
          onConfirm={() => remove(stockToRemove)}
          onCancel={() => setStockToRemove(null)}
        />
      )}
    </section>
  );
}

function ShoppingLists({ list, canManage, client, refresh }: { list: ShoppingList | null; canManage: boolean; client: ReturnType<typeof makeClient>; refresh: () => Promise<void> }) {
  async function generate() {
    await client.post('/api/shopping-lists/generate', {});
    await refresh();
  }
  return (
    <Panel title="Guild Shopping List">
      {canManage && <button className="primary inlineAction" onClick={generate}>Generate From Approved Shortages</button>}
      {list ? (
        <Table columns={['Item', 'Equivalent', 'Need']} rows={(list.items ?? []).map((i) => [i.itemName, `T${i.equivalentTier} Equivalent`, i.quantityNeeded])} />
      ) : <p className="muted">No shopping list generated yet.</p>}
    </Panel>
  );
}

function Crafting({ shoppingList }: { shoppingList: ShoppingList | null }) {
  return (
    <Panel title="Crafting Orders">
      <p className="muted">Recipe tables are included in the backend. The MVP shopping list below is ready to feed a material calculator endpoint.</p>
      {shoppingList && <Table columns={['Output', 'Qty']} rows={(shoppingList.items ?? []).map((i) => [equivalentItemLabel(i), i.quantityNeeded])} />}
    </Panel>
  );
}

function Members({ members, currentUser, canManage, client, refresh }: { members: MemberHistory[]; currentUser: User; canManage: boolean; client: ReturnType<typeof makeClient>; refresh: () => Promise<void> }) {
  if (!canManage) return <Panel title="Guild Roster"><p className="muted">Officer permission required.</p></Panel>;

  const [memberToRemove, setMemberToRemove] = useState<MemberHistory | null>(null);
  const roleValue = { 'Member': 1, 'Officer': 2, 'Admin': 3, 'Owner': 4 };
  const currentRoleVal = roleValue[currentUser.role] || 1;

  async function updateRole(memberId: number, newRole: string) {
    try {
      await client.patch(`/api/members/${memberId}/role`, { role: newRole });
      await refresh();
    } catch (err) {
      // client automatically sets error banner
    }
  }

  async function removeMember(member: MemberHistory) {
    try {
      await client.delete(`/api/members/${member.id}`);
      await refresh();
    } catch (err) {
      // client automatically sets error banner
    } finally {
      setMemberToRemove(null);
    }
  }

  return (
    <Panel title="Guild Roster">
      <Table 
        columns={['Player', 'Role', 'Requested', 'Approved', 'Actions']}
        rows={members.map((m) => {
          const mRoleVal = roleValue[m.role as keyof typeof roleValue] || 1;
          const canEdit = currentRoleVal > mRoleVal;
          
          const roleCell = canEdit ? (
            <select value={m.role} onChange={(e) => updateRole(m.id, e.target.value)}>
              {currentRoleVal > 1 && <option value="Member">Member</option>}
              {currentRoleVal > 2 && <option value="Officer">Officer</option>}
              {currentRoleVal > 3 && <option value="Admin">Admin</option>}
            </select>
          ) : m.role;

          return [
            m.playerName, 
            roleCell, 
            m.requested, 
            m.approved, 
            canEdit ? (
              <button 
                className="danger" 
                onClick={() => setMemberToRemove(m)}
                style={{ padding: '6px 12px', fontSize: '11px', fontWeight: 700 }}
              >
                Remove
              </button>
            ) : '—'
          ];
        })}
      />
      {memberToRemove && (
        <ConfirmDialog
          title="Remove Member"
          message={`Are you sure you want to remove ${memberToRemove.playerName} and delete their account entirely?`}
          confirmLabel="Yes, Remove"
          cancelLabel="No, Keep Them"
          onConfirm={() => removeMember(memberToRemove)}
          onCancel={() => setMemberToRemove(null)}
        />
      )}
    </Panel>
  );
}

function AdminSettings({ user }: { user: User }) {
  return (
    <Panel title="Admin Settings">
      <div className="settingsGrid">
        <div><strong>Current role</strong><span>{user.role}</span></div>
        <div><strong>Discord</strong><span>Outbox table ready</span></div>
        <div><strong>Database</strong><span>SQLite now, PostgreSQL path documented</span></div>
      </div>
    </Panel>
  );
}

function Stat({ label, value }: { label: string; value: string | number }) {
  return <article className="stat"><span>{label}</span><strong>{value}</strong></article>;
}

function Panel({ title, children }: { title: string; children: React.ReactNode }) {
  return <section className="panel"><h2>{title}</h2>{children}</section>;
}

function Table({ columns, rows }: { columns: string[]; rows: (string | number | React.ReactNode)[][] }) {
  return (
    <div className="tableWrap">
      <table>
        <thead><tr>{columns.map((c) => <th key={c}>{c}</th>)}</tr></thead>
        <tbody>{rows.map((row, index) => <tr key={index}>{row.map((cell, i) => <td key={i}>{cell}</td>)}</tr>)}</tbody>
      </table>
    </div>
  );
}

function RegearTable({ rows }: { rows: Regear[] }) {
  const [expandedId, setExpandedId] = useState<number | null>(null);

  return (
    <div className="tableWrap">
      <table>
        <thead>
          <tr>
            <th>Player</th>
            <th>Build</th>
            <th>Status</th>
            <th>Date</th>
          </tr>
        </thead>
        <tbody>
          {(rows ?? []).map((r) => {
            const isExpanded = expandedId === r.id;
            return (
              <React.Fragment key={r.id}>
                <tr
                  className="clickableRow"
                  onClick={() => setExpandedId(isExpanded ? null : r.id)}
                >
                  <td>{r.playerName}</td>
                  <td>{r.buildName}</td>
                  <td><StatusBadge value={r.status} /></td>
                  <td>{new Date(r.requestDate).toLocaleDateString()}</td>
                </tr>
                {isExpanded && (
                  <tr className="expandedRow">
                    <td colSpan={4}>
                      <div className="regearDetails">
                        {r.notes && (
                          <div className="regearDetailBlock">
                            <label>Notes</label>
                            <p>{r.notes}</p>
                          </div>
                        )}
                        {r.pickupLocation && isPickupLocationSet(r.pickupLocation) && (() => {
                          const { column, row } = parsePickupLocation(r.pickupLocation);
                          return (
                            <div className="regearDetailBlock">
                              <label>Pickup Location</label>
                              <div className="pickupLocationGrid" style={{ maxWidth: '200px' }}>
                                <div className="pickupLocationCell">
                                  <span>Column</span>
                                  <strong>{column || '—'}</strong>
                                </div>
                                <div className="pickupLocationCell">
                                  <span>Row</span>
                                  <strong>{row || '—'}</strong>
                                </div>
                              </div>
                            </div>
                          );
                        })()}
                        <RegearPreviews deathScreenshotUrl={r.deathScreenshotUrl} vodUrl={r.vodUrl} />
                      </div>
                    </td>
                  </tr>
                )}
              </React.Fragment>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

function BarList({ items }: { items: { label: string; value: number }[] }) {
  const max = Math.max(...items.map((i) => i.value), 1);
  return (
    <div className="barList">
      {items.map((item) => <div key={item.label}><span>{item.label}</span><div><i style={{ width: `${(item.value / max) * 100}%` }} /></div><b>{item.value}</b></div>)}
    </div>
  );
}

function StatusBadge({ value }: { value: string }) {
  const baseClass = value.split(' ')[0].toLowerCase();
  return <span className={`badge ${baseClass}`}>{value}</span>;
}

function makeClient(setError: (message: string) => void) {
  let token = '';
  async function request<T>(path: string, init: RequestInit = {}, options: { silentNotFound?: boolean } = {}): Promise<T> {
    setError('');
    let res: Response;
    try {
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...(init.headers as Record<string, string> ?? {})
      };
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
      res = await fetch(`${API_URL}${path}`, {
        ...init,
        headers
      });
    } catch (err) {
      const message = `Cannot reach API at ${API_URL}. Make sure the backend is running.`;
      setError(message);
      throw err;
    }
    if (!res.ok) {
      if (res.status === 404 && options.silentNotFound) {
        return null as T;
      }
      const body = await res.json().catch(() => ({ error: res.statusText }));
      const message = body.error ?? 'Request failed';
      setError(message);
      throw new Error(message);
    }
    return res.json();
  }
  return {
    setToken: (t: string) => { token = t; },
    get: <T,>(path: string) => request<T>(path),
    getOptional: <T,>(path: string) => request<T | null>(path, {}, { silentNotFound: true }),
    post: <T,>(path: string, body: unknown) => request<T>(path, { method: 'POST', body: JSON.stringify(body) }),
    patch: <T,>(path: string, body: unknown) => request<T>(path, { method: 'PATCH', body: JSON.stringify(body) }),
    delete: <T,>(path: string) => request<T>(path, { method: 'DELETE' })
  };
}

function updateAt<T>(items: T[], index: number, value: T) {
  return items.map((item, i) => (i === index ? value : item));
}

function loadScreenshot(event: React.ChangeEvent<HTMLInputElement>, setScreenshotUrl: (value: string) => void) {
  const file = event.target.files?.[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = () => {
    if (typeof reader.result === 'string') {
      setScreenshotUrl(reader.result);
    }
  };
  reader.readAsDataURL(file);
}

function itemLabel(item: { itemName: string; tier: number; enchantment: number }) {
  return `T${item.tier}.${item.enchantment} ${item.itemName}`;
}

function equivalentItemLabel(item: { itemName: string; equivalentTier: number }) {
  return `T${item.equivalentTier} Equivalent ${item.itemName}`;
}


function buildItemLabel(item: { itemName: string; tier: number; enchantment: number; slot?: string }) {
  if (item.slot === 'Food') {
    const power = consumablePowers.find((option) => option.enchantment === item.enchantment);
    return `${power?.label ?? `.${item.enchantment}`} ${item.itemName}`;
  }
  if (item.slot === 'Potion') {
    if (item.enchantment === 0) {
      return `T${item.tier} ${item.itemName}`;
    }
    return `T${item.tier}.${item.enchantment} ${item.itemName}`;
  }
  const power = buildPowers.find((option) => option.tier === item.tier && option.enchantment === item.enchantment);
  return `${power?.label ?? `T${item.tier}.${item.enchantment}`} ${item.itemName}`;
}

function silver(value: number) {
  return new Intl.NumberFormat('en-US', { notation: 'compact' }).format(value);
}

function getVodEmbedUrl(url: string): string | null {
  try {
    const u = new URL(url);
    // YouTube: youtube.com/watch?v=ID or youtu.be/ID
    if (u.hostname.includes('youtube.com') && u.searchParams.get('v')) {
      return `https://www.youtube.com/embed/${u.searchParams.get('v')}`;
    }
    if (u.hostname === 'youtu.be') {
      return `https://www.youtube.com/embed${u.pathname}`;
    }
    // Medal.tv: medal.tv/games/*/clips/ID or medal.tv/clips/ID
    if (u.hostname.includes('medal.tv')) {
      return url + (url.includes('?') ? '&' : '?') + 'autoplay=0';
    }
    // Twitch clips: clips.twitch.tv/SLUG or twitch.tv/*/clip/SLUG
    if (u.hostname === 'clips.twitch.tv') {
      const slug = u.pathname.replace(/^\//, '');
      return `https://clips.twitch.tv/embed?clip=${slug}&parent=${window.location.hostname}`;
    }
    if (u.hostname.includes('twitch.tv') && u.pathname.includes('/clip/')) {
      const slug = u.pathname.split('/clip/')[1];
      if (slug) return `https://clips.twitch.tv/embed?clip=${slug}&parent=${window.location.hostname}`;
    }
  } catch {
    // not a valid URL
  }
  return null;
}

function getVodProviderLabel(url: string): string {
  try {
    const u = new URL(url);
    if (u.hostname.includes('youtube') || u.hostname === 'youtu.be') return 'YouTube';
    if (u.hostname.includes('medal.tv')) return 'Medal';
    if (u.hostname.includes('twitch.tv')) return 'Twitch';
  } catch { /* ignore */ }
  return 'VOD';
}

function RegearPreviews({ deathScreenshotUrl, vodUrl }: { deathScreenshotUrl?: string; vodUrl?: string }) {
  const embedUrl = vodUrl ? getVodEmbedUrl(vodUrl) : null;
  const providerLabel = vodUrl ? getVodProviderLabel(vodUrl) : 'VOD';

  return (
    <>
      {deathScreenshotUrl && (
        <div className="screenshotPreviewContainer" style={{ marginTop: '10px' }}>
          <a href={deathScreenshotUrl} target="_blank" rel="noopener noreferrer" title="Click to view full image in a new tab">
            <img className="requestScreenshot" src={deathScreenshotUrl} alt="Death Screenshot" />
          </a>
          <div className="screenshotLink">
            <small>
              <a href={deathScreenshotUrl} target="_blank" rel="noopener noreferrer" style={{ color: 'var(--text-muted)', fontSize: '11px' }}>
                Open screenshot in new tab ↗
              </a>
            </small>
          </div>
        </div>
      )}
      {vodUrl && (
        <div className="vodContainer" style={{ marginTop: '12px' }}>
          {embedUrl ? (
            <div className="vodIframeContainer">
              <iframe className="vodIframe" src={embedUrl} title={`${providerLabel} VOD`} sandbox="allow-scripts allow-same-origin allow-popups" allowFullScreen />
            </div>
          ) : (
            <div className="vodFallback">
              <span>🎬</span>
              <a href={vodUrl} target="_blank" rel="noopener noreferrer">
                Watch VOD ↗
              </a>
            </div>
          )}
          <div className="vodLink">
            <small>
              <a href={vodUrl} target="_blank" rel="noopener noreferrer" style={{ color: 'var(--text-muted)', fontSize: '11px' }}>
                Open {providerLabel} in new tab ↗
              </a>
            </small>
          </div>
        </div>
      )}
    </>
  );
}

createRoot(document.getElementById('root')!).render(<App />);
